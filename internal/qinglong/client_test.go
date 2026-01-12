package qinglong

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestClient_TokenCachingAndAPIs(t *testing.T) {
	t.Parallel()

	var tokenHits int32
	var listHits int32
	var getHits int32
	var runHits int32
	var enableHits int32
	var disableHits int32
	var logHits int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/open/auth/token":
			atomic.AddInt32(&tokenHits, 1)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 200,
				"data": map[string]interface{}{
					"token":       "AT",
					"token_type":  "Bearer",
					"expiration":  time.Now().Add(1 * time.Hour).Unix(),
					"unexpected":  "ignored",
					"expiration2": nil,
				},
			})
			return

		case r.URL.Path == "/open/crons" && r.Method == http.MethodGet:
			atomic.AddInt32(&listHits, 1)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 200,
				"data": map[string]interface{}{
					"data": []map[string]interface{}{
						{"id": 1, "name": "a", "command": "c", "schedule": "* * * * *", "isDisabled": 0, "status": 0},
					},
					"total": 1,
				},
			})
			return

		case strings.HasPrefix(r.URL.Path, "/open/crons/") && r.Method == http.MethodGet:
			switch {
			case strings.HasSuffix(r.URL.Path, "/log"):
				atomic.AddInt32(&logHits, 1)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"code": 200,
					"data": "hello log",
				})
				return
			default:
				atomic.AddInt32(&getHits, 1)
				idStr := strings.TrimPrefix(r.URL.Path, "/open/crons/")
				id, _ := strconv.Atoi(idStr)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"code": 200,
					"data": map[string]interface{}{
						"id":         id,
						"name":       "job",
						"command":    "cmd",
						"schedule":   "0 0 * * *",
						"isDisabled": 0,
						"status":     0,
					},
				})
				return
			}

		case r.URL.Path == "/open/crons/run" && r.Method == http.MethodPut:
			atomic.AddInt32(&runHits, 1)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 200, "data": true})
			return

		case r.URL.Path == "/open/crons/enable" && r.Method == http.MethodPut:
			atomic.AddInt32(&enableHits, 1)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 200, "data": true})
			return

		case r.URL.Path == "/open/crons/disable" && r.Method == http.MethodPut:
			atomic.AddInt32(&disableHits, 1)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 200, "data": true})
			return
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}))
	t.Cleanup(srv.Close)

	c, err := NewClient(ClientConfig{
		BaseURL:      srv.URL,
		ClientID:     "id",
		ClientSecret: "sec",
	}, srv.Client())
	if err != nil {
		t.Fatalf("NewClient() error: %v", err)
	}

	ctx := context.Background()

	if _, err := c.ListCrons(ctx, ListCronsParams{SearchValue: "a", Page: 1, Size: 10}); err != nil {
		t.Fatalf("ListCrons() error: %v", err)
	}
	if _, err := c.GetCron(ctx, 1); err != nil {
		t.Fatalf("GetCron() error: %v", err)
	}
	if err := c.RunCrons(ctx, []int{1}); err != nil {
		t.Fatalf("RunCrons() error: %v", err)
	}
	if err := c.EnableCrons(ctx, []int{1}); err != nil {
		t.Fatalf("EnableCrons() error: %v", err)
	}
	if err := c.DisableCrons(ctx, []int{1}); err != nil {
		t.Fatalf("DisableCrons() error: %v", err)
	}
	if _, err := c.GetCronLog(ctx, 1); err != nil {
		t.Fatalf("GetCronLog() error: %v", err)
	}

	if atomic.LoadInt32(&tokenHits) != 1 {
		t.Fatalf("token hits = %d, want 1", tokenHits)
	}
	if atomic.LoadInt32(&listHits) != 1 {
		t.Fatalf("list hits = %d, want 1", listHits)
	}
	if atomic.LoadInt32(&getHits) != 1 {
		t.Fatalf("get hits = %d, want 1", getHits)
	}
	if atomic.LoadInt32(&runHits) != 1 {
		t.Fatalf("run hits = %d, want 1", runHits)
	}
	if atomic.LoadInt32(&enableHits) != 1 {
		t.Fatalf("enable hits = %d, want 1", enableHits)
	}
	if atomic.LoadInt32(&disableHits) != 1 {
		t.Fatalf("disable hits = %d, want 1", disableHits)
	}
	if atomic.LoadInt32(&logHits) != 1 {
		t.Fatalf("log hits = %d, want 1", logHits)
	}
}

func TestClient_ConcurrentTokenRefresh(t *testing.T) {
	t.Parallel()

	var tokenHits int32
	var listHits int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/open/auth/token":
			atomic.AddInt32(&tokenHits, 1)
			time.Sleep(20 * time.Millisecond)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 200,
				"data": map[string]interface{}{
					"token":      "AT",
					"token_type": "Bearer",
					"expiration": time.Now().Add(1 * time.Hour).Unix(),
				},
			})
			return

		case r.URL.Path == "/open/crons" && r.Method == http.MethodGet:
			atomic.AddInt32(&listHits, 1)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 200,
				"data": map[string]interface{}{
					"data":  []map[string]interface{}{},
					"total": 0,
				},
			})
			return

		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}))
	t.Cleanup(srv.Close)

	c, err := NewClient(ClientConfig{
		BaseURL:      srv.URL,
		ClientID:     "id",
		ClientSecret: "sec",
	}, srv.Client())
	if err != nil {
		t.Fatalf("NewClient() error: %v", err)
	}

	ctx := context.Background()

	const n = 10
	var wg sync.WaitGroup
	wg.Add(n)
	errCh := make(chan error, n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			if _, err := c.ListCrons(ctx, ListCronsParams{SearchValue: "a", Page: 1, Size: 10}); err != nil {
				errCh <- err
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			t.Fatalf("ListCrons() error: %v", err)
		}
	}

	if atomic.LoadInt32(&tokenHits) != 1 {
		t.Fatalf("token hits = %d, want 1", tokenHits)
	}
	if atomic.LoadInt32(&listHits) != n {
		t.Fatalf("list hits = %d, want %d", listHits, n)
	}
}

func TestNewClient_InvalidConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cfg  ClientConfig
	}{
		{name: "empty base url", cfg: ClientConfig{BaseURL: "", ClientID: "id", ClientSecret: "sec"}},
		{name: "bad scheme", cfg: ClientConfig{BaseURL: "ftp://example.com", ClientID: "id", ClientSecret: "sec"}},
		{name: "missing host", cfg: ClientConfig{BaseURL: "http://", ClientID: "id", ClientSecret: "sec"}},
		{name: "empty client id", cfg: ClientConfig{BaseURL: "http://example.com", ClientID: "", ClientSecret: "sec"}},
		{name: "empty client secret", cfg: ClientConfig{BaseURL: "http://example.com", ClientID: "id", ClientSecret: ""}},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if _, err := NewClient(tc.cfg, &http.Client{}); err == nil {
				t.Fatalf("NewClient() error = nil, want not nil")
			}
		})
	}
}

func TestNewClient_NormalizesBaseURL(t *testing.T) {
	t.Parallel()

	c, err := NewClient(ClientConfig{
		BaseURL:      "http://example.com/",
		ClientID:     "id",
		ClientSecret: "sec",
	}, &http.Client{})
	if err != nil {
		t.Fatalf("NewClient() error: %v", err)
	}
	if c.cfg.BaseURL != "http://example.com" {
		t.Fatalf("BaseURL = %q, want %q", c.cfg.BaseURL, "http://example.com")
	}
}
