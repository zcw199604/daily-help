package wecom

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
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
