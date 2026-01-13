package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"github.com/zcw199604/wecom-home-ops/internal/app"
	"github.com/zcw199604/wecom-home-ops/internal/config"
	"github.com/zcw199604/wecom-home-ops/internal/wecom"
)

func main() {
	var configPath string
	var wecomSyncMenu bool
	flag.StringVar(&configPath, "config", "config.yaml", "配置文件路径（YAML）")
	flag.BoolVar(&wecomSyncMenu, "wecom-sync-menu", false, "同步企业微信应用自定义菜单（menu/create）后退出")
	flag.Parse()

	startedAt := time.Now()

	bootstrapLogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(bootstrapLogger)

	cfg, err := config.Load(configPath)
	if err != nil {
		slog.Error("加载配置失败", "path", configPath, "error", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.Log.Level.ToSlogLevel(),
	}))
	slog.SetDefault(logger)

	if wecomSyncMenu {
		httpClient := &http.Client{
			Timeout: cfg.Server.HTTPClientTimeout.ToDuration(),
		}
		wecomClient := wecom.NewClient(wecom.ClientConfig{
			APIBaseURL: cfg.WeCom.APIBaseURL,
			CorpID:     cfg.WeCom.CorpID,
			AgentID:    cfg.WeCom.AgentID,
			Secret:     cfg.WeCom.Secret,
		}, httpClient)

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		if err := wecomClient.CreateMenu(ctx, wecom.DefaultMenu()); err != nil {
			slog.Error("企业微信自定义菜单同步失败", "error", err)
			os.Exit(1)
		}
		slog.Info("企业微信自定义菜单同步成功")
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	server, err := app.NewServer(cfg)
	if err != nil {
		slog.Error("初始化失败", "error", err)
		os.Exit(1)
	}

	listener, err := net.Listen("tcp", cfg.Server.ListenAddr)
	if err != nil {
		slog.Error("HTTP 端口监听失败", "listen_addr", cfg.Server.ListenAddr, "error", err)
		os.Exit(1)
	}
	readyAt := time.Now()

	go func() {
		if err := server.Serve(listener); err != nil {
			slog.Error("HTTP 服务启动失败", "error", err)
			stop()
		}
	}()

	go func() {
		notifyCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		sendStartupSuccessNotification(notifyCtx, cfg, configPath, listener.Addr().String(), startedAt, readyAt)
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("HTTP 服务关闭失败", "error", err)
	}
}

func sendStartupSuccessNotification(ctx context.Context, cfg config.Config, configPath string, listenerAddr string, startedAt, readyAt time.Time) {
	userIDs := uniqueNonEmpty(cfg.Auth.AllowedUserIDs)
	if len(userIDs) == 0 {
		return
	}

	httpClient := &http.Client{
		Timeout: cfg.Server.HTTPClientTimeout.ToDuration(),
	}
	wecomClient := wecom.NewClient(wecom.ClientConfig{
		APIBaseURL: cfg.WeCom.APIBaseURL,
		CorpID:     cfg.WeCom.CorpID,
		AgentID:    cfg.WeCom.AgentID,
		Secret:     cfg.WeCom.Secret,
	}, httpClient)

	content := buildStartupSuccessMessage(cfg, configPath, listenerAddr, startedAt, readyAt)
	toUser := strings.Join(userIDs, "|")

	if err := wecomClient.SendText(ctx, wecom.TextMessage{
		ToUser:  toUser,
		Content: content,
	}); err != nil {
		slog.Error("启动成功通知发送失败", "error", err, "users_count", len(userIDs))
		return
	}
	slog.Info("启动成功通知已发送", "users_count", len(userIDs))
}

func buildStartupSuccessMessage(cfg config.Config, configPath string, listenerAddr string, startedAt, readyAt time.Time) string {
	hostname, _ := os.Hostname()
	pid := os.Getpid()

	var b strings.Builder
	b.WriteString("wecom-home-ops 启动成功\n\n")
	fmt.Fprintf(&b, "时间: %s\n", readyAt.Format(time.RFC3339))
	fmt.Fprintf(&b, "耗时: %s\n", readyAt.Sub(startedAt).Truncate(time.Millisecond))
	if strings.TrimSpace(hostname) != "" {
		fmt.Fprintf(&b, "主机: %s\n", hostname)
	}
	fmt.Fprintf(&b, "PID: %d\n", pid)
	fmt.Fprintf(&b, "Go: %s\n", runtime.Version())
	if bi := buildInfoSummary(); bi != "" {
		fmt.Fprintf(&b, "Build: %s\n", bi)
	}
	fmt.Fprintf(&b, "配置: %s\n", configPath)

	fmt.Fprintf(&b, "\n监听:\n- 配置: %s\n- 实际: %s\n", cfg.Server.ListenAddr, listenerAddr)
	b.WriteString("\n入口:\n- 回调: /wecom/callback\n- 健康: /healthz\n- 就绪: /readyz\n")

	baseURL := strings.TrimSpace(cfg.Server.BaseURL)
	if baseURL != "" {
		fmt.Fprintf(&b, "\nBaseURL: %s\n", baseURL)
		for _, p := range []string{"/wecom/callback", "/healthz", "/readyz"} {
			if u := joinURLPath(baseURL, p); u != "" {
				fmt.Fprintf(&b, "- %s\n", u)
			}
		}
	} else {
		b.WriteString("\nBaseURL: <未配置>\n")
	}

	unraidEnabled := strings.TrimSpace(cfg.Unraid.Endpoint) != "" && strings.TrimSpace(cfg.Unraid.APIKey) != ""
	b.WriteString("\n后端:\n")
	if unraidEnabled {
		b.WriteString("- Unraid: 已启用\n")
	} else {
		b.WriteString("- Unraid: 未启用\n")
	}

	if len(cfg.Qinglong.Instances) == 0 {
		b.WriteString("- Qinglong: 未启用\n")
	} else {
		var instances []string
		for _, ins := range cfg.Qinglong.Instances {
			id := strings.TrimSpace(ins.ID)
			name := strings.TrimSpace(ins.Name)
			if id != "" && name != "" {
				instances = append(instances, id+"("+name+")")
				continue
			}
			if id != "" {
				instances = append(instances, id)
				continue
			}
			if name != "" {
				instances = append(instances, name)
			}
		}
		fmt.Fprintf(&b, "- Qinglong: %d 个实例", len(cfg.Qinglong.Instances))
		if len(instances) > 0 {
			fmt.Fprintf(&b, "（%s）", strings.Join(instances, ", "))
		}
		b.WriteString("\n")
	}

	return b.String()
}

func buildInfoSummary() string {
	bi, ok := debug.ReadBuildInfo()
	if !ok || bi == nil {
		return ""
	}

	version := strings.TrimSpace(bi.Main.Version)
	if version == "" || version == "(devel)" {
		version = ""
	}

	var revision string
	var modified string
	var buildTime string
	for _, s := range bi.Settings {
		switch s.Key {
		case "vcs.revision":
			revision = strings.TrimSpace(s.Value)
		case "vcs.modified":
			modified = strings.TrimSpace(s.Value)
		case "vcs.time":
			buildTime = strings.TrimSpace(s.Value)
		}
	}

	if revision != "" && len(revision) > 12 {
		revision = revision[:12]
	}

	var parts []string
	if version != "" {
		parts = append(parts, version)
	}
	if revision != "" {
		parts = append(parts, revision)
	}
	if modified == "true" {
		parts = append(parts, "dirty")
	}
	if buildTime != "" {
		parts = append(parts, buildTime)
	}

	return strings.Join(parts, " ")
}

func joinURLPath(base string, p string) string {
	base = strings.TrimSpace(base)
	p = strings.TrimSpace(p)
	if base == "" || p == "" {
		return ""
	}
	u, err := url.JoinPath(base, strings.TrimPrefix(p, "/"))
	if err != nil {
		return ""
	}
	return u
}

func uniqueNonEmpty(ss []string) []string {
	if len(ss) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(ss))
	out := make([]string, 0, len(ss))
	for _, s := range ss {
		v := strings.TrimSpace(s)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}
