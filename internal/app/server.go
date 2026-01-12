package app

// server.go 负责装配依赖并启动 HTTP 路由（企业微信回调入口等）。
import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"daily-help/internal/config"
	"daily-help/internal/core"
	"daily-help/internal/qinglong"
	"daily-help/internal/unraid"
	"daily-help/internal/wecom"
)

type Server struct {
	cfg    config.Config
	server *http.Server
}

func NewServer(cfg config.Config) (*Server, error) {
	httpClient := &http.Client{
		Timeout: 15 * time.Second,
	}

	wecomClient := wecom.NewClient(wecom.ClientConfig{
		APIBaseURL: cfg.WeCom.APIBaseURL,
		CorpID:     cfg.WeCom.CorpID,
		AgentID:    cfg.WeCom.AgentID,
		Secret:     cfg.WeCom.Secret,
	}, httpClient)

	stateStore := core.NewStateStore(30 * time.Minute)

	var providers []core.ServiceProvider

	if cfg.Unraid.Endpoint != "" && cfg.Unraid.APIKey != "" {
		unraidClient := unraid.NewClient(unraid.ClientConfig{
			Endpoint:            cfg.Unraid.Endpoint,
			APIKey:              cfg.Unraid.APIKey,
			Origin:              cfg.Unraid.Origin,
			ForceUpdateMutation: cfg.Unraid.ForceUpdateMutation,
		}, httpClient)
		providers = append(providers, unraid.NewProvider(unraid.ProviderDeps{
			WeCom:  wecomClient,
			Client: unraidClient,
			State:  stateStore,
		}))
	}

	if len(cfg.Qinglong.Instances) > 0 {
		var instances []qinglong.Instance
		for _, ins := range cfg.Qinglong.Instances {
			client, err := qinglong.NewClient(qinglong.ClientConfig{
				BaseURL:      ins.BaseURL,
				ClientID:     ins.ClientID,
				ClientSecret: ins.ClientSecret,
			}, httpClient)
			if err != nil {
				return nil, err
			}
			instances = append(instances, qinglong.Instance{
				ID:     ins.ID,
				Name:   ins.Name,
				Client: client,
			})
		}
		providers = append(providers, qinglong.NewProvider(qinglong.ProviderDeps{
			WeCom:     wecomClient,
			State:     stateStore,
			Instances: instances,
		}))
	}

	router := core.NewRouter(core.RouterDeps{
		WeCom:         wecomClient,
		AllowedUserID: make(map[string]struct{}),
		Providers:     providers,
		State:         stateStore,
	})
	for _, id := range cfg.Auth.AllowedUserIDs {
		router.AllowedUserID[id] = struct{}{}
	}

	crypto, err := wecom.NewCrypto(wecom.CryptoConfig{
		Token:          cfg.WeCom.Token,
		EncodingAESKey: cfg.WeCom.EncodingAESKey,
		ReceiverID:     cfg.WeCom.CorpID,
	})
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.Handle("GET /wecom/callback", wecom.NewCallbackVerifyHandler(crypto))
	mux.Handle("POST /wecom/callback", wecom.NewCallbackHandler(wecom.CallbackDeps{
		Crypto: crypto,
		Core:   router,
	}))

	s := &http.Server{
		Addr:              cfg.Server.ListenAddr,
		Handler:           withRequestLogging(mux),
		ReadHeaderTimeout: 10 * time.Second,
	}

	return &Server{
		cfg:    cfg,
		server: s,
	}, nil
}

func (s *Server) Start() error {
	slog.Info("HTTP 服务启动", "listen_addr", s.cfg.Server.ListenAddr)
	err := s.server.ListenAndServe()
	if err == nil || errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return fmt.Errorf("listen and serve: %w", err)
}

func (s *Server) Shutdown(ctx context.Context) error {
	slog.Info("HTTP 服务关闭中")
	return s.server.Shutdown(ctx)
}

func withRequestLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		slog.Info("请求完成",
			"method", r.Method,
			"path", r.URL.Path,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}
