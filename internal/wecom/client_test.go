package wecom

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestClient_SendText_UsesCachedAccessToken(t *testing.T) {
	t.Parallel()

	var getTokenHits int32
	var sendHits int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/gettoken":
			atomic.AddInt32(&getTokenHits, 1)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"errcode":      0,
				"errmsg":       "ok",
				"access_token": "AT",
				"expires_in":   7200,
			})
			return
		case "/message/send":
			atomic.AddInt32(&sendHits, 1)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"errcode": 0,
				"errmsg":  "ok",
			})
			return
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}))
	t.Cleanup(srv.Close)

	c := NewClient(ClientConfig{
		APIBaseURL: srv.URL,
		CorpID:     "ww",
		AgentID:    1,
		Secret:     "sec",
	}, srv.Client())

	ctx := context.Background()
	if err := c.SendText(ctx, TextMessage{ToUser: "u", Content: "a"}); err != nil {
		t.Fatalf("SendText(1) error: %v", err)
	}
	if err := c.SendText(ctx, TextMessage{ToUser: "u", Content: "b"}); err != nil {
		t.Fatalf("SendText(2) error: %v", err)
	}

	if atomic.LoadInt32(&getTokenHits) != 1 {
		t.Fatalf("gettoken hits = %d, want 1", getTokenHits)
	}
	if atomic.LoadInt32(&sendHits) != 2 {
		t.Fatalf("message/send hits = %d, want 2", sendHits)
	}
}

func TestClient_SendText_ConcurrentTokenRefresh(t *testing.T) {
	t.Parallel()

	var getTokenHits int32
	var sendHits int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/gettoken":
			atomic.AddInt32(&getTokenHits, 1)
			time.Sleep(20 * time.Millisecond)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"errcode":      0,
				"errmsg":       "ok",
				"access_token": "AT",
				"expires_in":   7200,
			})
			return
		case "/message/send":
			atomic.AddInt32(&sendHits, 1)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"errcode": 0,
				"errmsg":  "ok",
			})
			return
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}))
	t.Cleanup(srv.Close)

	c := NewClient(ClientConfig{
		APIBaseURL: srv.URL,
		CorpID:     "ww",
		AgentID:    1,
		Secret:     "sec",
	}, srv.Client())

	ctx := context.Background()

	const n = 10
	var wg sync.WaitGroup
	wg.Add(n)
	errCh := make(chan error, n)
	for i := 0; i < n; i++ {
		i := i
		go func() {
			defer wg.Done()
			if err := c.SendText(ctx, TextMessage{ToUser: "u", Content: "msg-" + string(rune('a'+i))}); err != nil {
				errCh <- err
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			t.Fatalf("SendText() error: %v", err)
		}
	}

	if atomic.LoadInt32(&getTokenHits) != 1 {
		t.Fatalf("gettoken hits = %d, want 1", getTokenHits)
	}
	if atomic.LoadInt32(&sendHits) != n {
		t.Fatalf("message/send hits = %d, want %d", sendHits, n)
	}
}
