package wecom

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

type testCoreHandler struct {
	mu    sync.Mutex
	calls int
	err   error
}

func (h *testCoreHandler) HandleMessage(_ context.Context, _ IncomingMessage) error {
	h.mu.Lock()
	h.calls++
	h.mu.Unlock()
	return h.err
}

func (h *testCoreHandler) Calls() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.calls
}

func TestCallbackHandler_BodyTooLarge(t *testing.T) {
	t.Parallel()

	crypto := mustTestCrypto(t, "t", "ww123")
	core := &testCoreHandler{}
	deduper := NewDeduper(10 * time.Minute)
	t.Cleanup(deduper.Close)

	h := NewCallbackHandler(CallbackDeps{
		Crypto:       crypto,
		Core:         core,
		Deduper:      deduper,
		MaxBodyBytes: 16,
	})

	req := httptest.NewRequest(http.MethodPost, "/wecom/callback?msg_signature=x&timestamp=1&nonce=1", strings.NewReader(strings.Repeat("a", 64)))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusRequestEntityTooLarge)
	}
}

func TestCallbackHandler_MaxBodyBytes_DefaultWhenZero(t *testing.T) {
	t.Parallel()

	crypto := mustTestCrypto(t, "t", "ww123")
	core := &testCoreHandler{}

	h := NewCallbackHandler(CallbackDeps{
		Crypto:       crypto,
		Core:         core,
		MaxBodyBytes: 0,
	})

	req := httptest.NewRequest(http.MethodPost, "/wecom/callback?msg_signature=x&timestamp=1&nonce=1", strings.NewReader(strings.Repeat("a", (1<<20)+1)))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusRequestEntityTooLarge)
	}
}

func TestCallbackHandler_DedupByMsgID(t *testing.T) {
	t.Parallel()

	token := "test-token"
	crypto := mustTestCrypto(t, token, "ww123")
	core := &testCoreHandler{}
	deduper := NewDeduper(10 * time.Minute)
	t.Cleanup(deduper.Close)

	h := NewCallbackHandler(CallbackDeps{
		Crypto:  crypto,
		Core:    core,
		Deduper: deduper,
	})

	plain := []byte("<xml>" +
		"<ToUserName><![CDATA[to]]></ToUserName>" +
		"<FromUserName><![CDATA[user]]></FromUserName>" +
		"<CreateTime>1700000000</CreateTime>" +
		"<MsgType><![CDATA[text]]></MsgType>" +
		"<Content><![CDATA[菜单]]></Content>" +
		"<MsgId>12345</MsgId>" +
		"</xml>")
	encrypted := mustEncrypt(t, crypto, plain)

	timestamp := "1700000001"
	nonce := "nonce"
	sig := signature(token, timestamp, nonce, encrypted)

	body := []byte("<xml>" +
		"<ToUserName><![CDATA[to]]></ToUserName>" +
		"<Encrypt><![CDATA[" + encrypted + "]]></Encrypt>" +
		"</xml>")

	req1 := httptest.NewRequest(http.MethodPost, "/wecom/callback?msg_signature="+sig+"&timestamp="+timestamp+"&nonce="+nonce, bytes.NewReader(body))
	w1 := httptest.NewRecorder()
	h.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("first status = %d, want %d", w1.Code, http.StatusOK)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/wecom/callback?msg_signature="+sig+"&timestamp="+timestamp+"&nonce="+nonce, bytes.NewReader(body))
	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("second status = %d, want %d", w2.Code, http.StatusOK)
	}

	if got := core.Calls(); got != 1 {
		t.Fatalf("core calls = %d, want 1", got)
	}
}

func TestCallbackHandler_DedupByTaskID(t *testing.T) {
	t.Parallel()

	token := "test-token"
	crypto := mustTestCrypto(t, token, "ww123")
	core := &testCoreHandler{}
	deduper := NewDeduper(10 * time.Minute)
	t.Cleanup(deduper.Close)

	h := NewCallbackHandler(CallbackDeps{
		Crypto:  crypto,
		Core:    core,
		Deduper: deduper,
	})

	plain := []byte("<xml>" +
		"<ToUserName><![CDATA[to]]></ToUserName>" +
		"<FromUserName><![CDATA[user]]></FromUserName>" +
		"<CreateTime>1700000000</CreateTime>" +
		"<MsgType><![CDATA[text]]></MsgType>" +
		"<Content><![CDATA[菜单]]></Content>" +
		"<TaskId>t1</TaskId>" +
		"</xml>")
	encrypted := mustEncrypt(t, crypto, plain)

	timestamp := "1700000001"
	nonce := "nonce"
	sig := signature(token, timestamp, nonce, encrypted)

	body := []byte("<xml>" +
		"<ToUserName><![CDATA[to]]></ToUserName>" +
		"<Encrypt><![CDATA[" + encrypted + "]]></Encrypt>" +
		"</xml>")

	req1 := httptest.NewRequest(http.MethodPost, "/wecom/callback?msg_signature="+sig+"&timestamp="+timestamp+"&nonce="+nonce, bytes.NewReader(body))
	w1 := httptest.NewRecorder()
	h.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("first status = %d, want %d", w1.Code, http.StatusOK)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/wecom/callback?msg_signature="+sig+"&timestamp="+timestamp+"&nonce="+nonce, bytes.NewReader(body))
	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("second status = %d, want %d", w2.Code, http.StatusOK)
	}

	if got := core.Calls(); got != 1 {
		t.Fatalf("core calls = %d, want 1", got)
	}
}

func TestCallbackHandler_DedupByPlainHashWhenNoIDs(t *testing.T) {
	t.Parallel()

	token := "test-token"
	crypto := mustTestCrypto(t, token, "ww123")
	core := &testCoreHandler{}
	deduper := NewDeduper(10 * time.Minute)
	t.Cleanup(deduper.Close)

	h := NewCallbackHandler(CallbackDeps{
		Crypto:  crypto,
		Core:    core,
		Deduper: deduper,
	})

	plain := []byte("<xml>" +
		"<ToUserName><![CDATA[to]]></ToUserName>" +
		"<FromUserName><![CDATA[user]]></FromUserName>" +
		"<CreateTime>1700000000</CreateTime>" +
		"<MsgType><![CDATA[text]]></MsgType>" +
		"<Content><![CDATA[菜单]]></Content>" +
		"</xml>")
	encrypted := mustEncrypt(t, crypto, plain)

	timestamp := "1700000001"
	nonce := "nonce"
	sig := signature(token, timestamp, nonce, encrypted)

	body := []byte("<xml>" +
		"<ToUserName><![CDATA[to]]></ToUserName>" +
		"<Encrypt><![CDATA[" + encrypted + "]]></Encrypt>" +
		"</xml>")

	req1 := httptest.NewRequest(http.MethodPost, "/wecom/callback?msg_signature="+sig+"&timestamp="+timestamp+"&nonce="+nonce, bytes.NewReader(body))
	w1 := httptest.NewRecorder()
	h.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("first status = %d, want %d", w1.Code, http.StatusOK)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/wecom/callback?msg_signature="+sig+"&timestamp="+timestamp+"&nonce="+nonce, bytes.NewReader(body))
	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("second status = %d, want %d", w2.Code, http.StatusOK)
	}

	if got := core.Calls(); got != 1 {
		t.Fatalf("core calls = %d, want 1", got)
	}
}

func TestCallbackHandler_DeduperNil_NoDedup(t *testing.T) {
	t.Parallel()

	token := "test-token"
	crypto := mustTestCrypto(t, token, "ww123")
	core := &testCoreHandler{}

	h := NewCallbackHandler(CallbackDeps{
		Crypto: crypto,
		Core:   core,
	})

	plain := []byte("<xml>" +
		"<ToUserName><![CDATA[to]]></ToUserName>" +
		"<FromUserName><![CDATA[user]]></FromUserName>" +
		"<CreateTime>1700000000</CreateTime>" +
		"<MsgType><![CDATA[text]]></MsgType>" +
		"<Content><![CDATA[菜单]]></Content>" +
		"<MsgId>12345</MsgId>" +
		"</xml>")
	encrypted := mustEncrypt(t, crypto, plain)

	timestamp := "1700000001"
	nonce := "nonce"
	sig := signature(token, timestamp, nonce, encrypted)

	body := []byte("<xml>" +
		"<ToUserName><![CDATA[to]]></ToUserName>" +
		"<Encrypt><![CDATA[" + encrypted + "]]></Encrypt>" +
		"</xml>")

	req1 := httptest.NewRequest(http.MethodPost, "/wecom/callback?msg_signature="+sig+"&timestamp="+timestamp+"&nonce="+nonce, bytes.NewReader(body))
	w1 := httptest.NewRecorder()
	h.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("first status = %d, want %d", w1.Code, http.StatusOK)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/wecom/callback?msg_signature="+sig+"&timestamp="+timestamp+"&nonce="+nonce, bytes.NewReader(body))
	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("second status = %d, want %d", w2.Code, http.StatusOK)
	}

	if got := core.Calls(); got != 2 {
		t.Fatalf("core calls = %d, want 2", got)
	}
}

func TestCallbackHandler_InvalidSignature(t *testing.T) {
	t.Parallel()

	token := "t"
	crypto := mustTestCrypto(t, token, "ww123")
	core := &testCoreHandler{}

	h := NewCallbackHandler(CallbackDeps{
		Crypto: crypto,
		Core:   core,
	})

	plain := []byte("<xml>" +
		"<ToUserName><![CDATA[to]]></ToUserName>" +
		"<FromUserName><![CDATA[user]]></FromUserName>" +
		"<CreateTime>1700000000</CreateTime>" +
		"<MsgType><![CDATA[text]]></MsgType>" +
		"<Content><![CDATA[菜单]]></Content>" +
		"<MsgId>12345</MsgId>" +
		"</xml>")
	encrypted := mustEncrypt(t, crypto, plain)

	body := []byte("<xml>" +
		"<ToUserName><![CDATA[to]]></ToUserName>" +
		"<Encrypt><![CDATA[" + encrypted + "]]></Encrypt>" +
		"</xml>")

	// 故意传入错误签名，确保命中 VerifySignature 分支。
	req := httptest.NewRequest(http.MethodPost, "/wecom/callback?msg_signature=bad&timestamp=1&nonce=1", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
	if got := core.Calls(); got != 0 {
		t.Fatalf("core calls = %d, want 0", got)
	}
}

func TestCallbackHandler_MissingSignatureParam(t *testing.T) {
	t.Parallel()

	token := "t"
	crypto := mustTestCrypto(t, token, "ww123")
	core := &testCoreHandler{}

	h := NewCallbackHandler(CallbackDeps{
		Crypto: crypto,
		Core:   core,
	})

	plain := []byte("<xml>" +
		"<ToUserName><![CDATA[to]]></ToUserName>" +
		"<FromUserName><![CDATA[user]]></FromUserName>" +
		"<CreateTime>1700000000</CreateTime>" +
		"<MsgType><![CDATA[text]]></MsgType>" +
		"<Content><![CDATA[菜单]]></Content>" +
		"<MsgId>12345</MsgId>" +
		"</xml>")
	encrypted := mustEncrypt(t, crypto, plain)

	timestamp := "1"
	nonce := "1"

	body := []byte("<xml>" +
		"<ToUserName><![CDATA[to]]></ToUserName>" +
		"<Encrypt><![CDATA[" + encrypted + "]]></Encrypt>" +
		"</xml>")

	req := httptest.NewRequest(http.MethodPost, "/wecom/callback?timestamp="+timestamp+"&nonce="+nonce, bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
	if got := core.Calls(); got != 0 {
		t.Fatalf("core calls = %d, want 0", got)
	}
}

func TestCallbackHandler_MissingEncrypt(t *testing.T) {
	t.Parallel()

	crypto := mustTestCrypto(t, "t", "ww123")
	core := &testCoreHandler{}

	h := NewCallbackHandler(CallbackDeps{
		Crypto: crypto,
		Core:   core,
	})

	req := httptest.NewRequest(http.MethodPost, "/wecom/callback?msg_signature=x&timestamp=1&nonce=1", strings.NewReader("<xml><ToUserName><![CDATA[to]]></ToUserName></xml>"))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	if got := core.Calls(); got != 0 {
		t.Fatalf("core calls = %d, want 0", got)
	}
}

func TestCallbackHandler_BadEnvelopeXML(t *testing.T) {
	t.Parallel()

	crypto := mustTestCrypto(t, "t", "ww123")
	core := &testCoreHandler{}

	h := NewCallbackHandler(CallbackDeps{
		Crypto: crypto,
		Core:   core,
	})

	req := httptest.NewRequest(http.MethodPost, "/wecom/callback?msg_signature=x&timestamp=1&nonce=1", strings.NewReader("not-xml"))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	if got := core.Calls(); got != 0 {
		t.Fatalf("core calls = %d, want 0", got)
	}
}

func TestCallbackHandler_DecryptFailure(t *testing.T) {
	t.Parallel()

	token := "t"
	crypto := mustTestCrypto(t, token, "ww123")
	core := &testCoreHandler{}

	h := NewCallbackHandler(CallbackDeps{
		Crypto: crypto,
		Core:   core,
	})

	encrypted := "YWI="
	timestamp := "1"
	nonce := "1"
	sig := signature(token, timestamp, nonce, encrypted)

	body := []byte("<xml>" +
		"<ToUserName><![CDATA[to]]></ToUserName>" +
		"<Encrypt><![CDATA[" + encrypted + "]]></Encrypt>" +
		"</xml>")

	req := httptest.NewRequest(http.MethodPost, "/wecom/callback?msg_signature="+sig+"&timestamp="+timestamp+"&nonce="+nonce, bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
	if got := core.Calls(); got != 0 {
		t.Fatalf("core calls = %d, want 0", got)
	}
}

func TestCallbackHandler_BadDecryptedXML(t *testing.T) {
	t.Parallel()

	token := "test-token"
	crypto := mustTestCrypto(t, token, "ww123")
	core := &testCoreHandler{}

	h := NewCallbackHandler(CallbackDeps{
		Crypto: crypto,
		Core:   core,
	})

	plain := []byte("not xml")
	encrypted := mustEncrypt(t, crypto, plain)

	timestamp := "1700000001"
	nonce := "nonce"
	sig := signature(token, timestamp, nonce, encrypted)

	body := []byte("<xml>" +
		"<ToUserName><![CDATA[to]]></ToUserName>" +
		"<Encrypt><![CDATA[" + encrypted + "]]></Encrypt>" +
		"</xml>")

	req := httptest.NewRequest(http.MethodPost, "/wecom/callback?msg_signature="+sig+"&timestamp="+timestamp+"&nonce="+nonce, bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	if got := core.Calls(); got != 0 {
		t.Fatalf("core calls = %d, want 0", got)
	}
}

func TestCallbackHandler_MaxBodyBytes_BoundaryExactlyLimit(t *testing.T) {
	t.Parallel()

	token := "test-token"
	crypto := mustTestCrypto(t, token, "ww123")
	core := &testCoreHandler{}

	plain := []byte("<xml>" +
		"<ToUserName><![CDATA[to]]></ToUserName>" +
		"<FromUserName><![CDATA[user]]></FromUserName>" +
		"<CreateTime>1700000000</CreateTime>" +
		"<MsgType><![CDATA[text]]></MsgType>" +
		"<Content><![CDATA[菜单]]></Content>" +
		"<MsgId>12345</MsgId>" +
		"</xml>")
	encrypted := mustEncrypt(t, crypto, plain)

	timestamp := "1700000001"
	nonce := "nonce"
	sig := signature(token, timestamp, nonce, encrypted)

	body := []byte("<xml>" +
		"<ToUserName><![CDATA[to]]></ToUserName>" +
		"<Encrypt><![CDATA[" + encrypted + "]]></Encrypt>" +
		"</xml>")

	h := NewCallbackHandler(CallbackDeps{
		Crypto:       crypto,
		Core:         core,
		MaxBodyBytes: int64(len(body)),
	})

	req := httptest.NewRequest(http.MethodPost, "/wecom/callback?msg_signature="+sig+"&timestamp="+timestamp+"&nonce="+nonce, bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if got := core.Calls(); got != 1 {
		t.Fatalf("core calls = %d, want 1", got)
	}
}

func TestCallbackHandler_CoreErrorStill200(t *testing.T) {
	t.Parallel()

	token := "test-token"
	crypto := mustTestCrypto(t, token, "ww123")
	core := &testCoreHandler{err: errors.New("boom")}
	deduper := NewDeduper(10 * time.Minute)
	t.Cleanup(deduper.Close)

	h := NewCallbackHandler(CallbackDeps{
		Crypto:  crypto,
		Core:    core,
		Deduper: deduper,
	})

	plain := []byte("<xml>" +
		"<ToUserName><![CDATA[to]]></ToUserName>" +
		"<FromUserName><![CDATA[user]]></FromUserName>" +
		"<CreateTime>1700000000</CreateTime>" +
		"<MsgType><![CDATA[text]]></MsgType>" +
		"<Content><![CDATA[菜单]]></Content>" +
		"<MsgId>12345</MsgId>" +
		"</xml>")
	encrypted := mustEncrypt(t, crypto, plain)

	timestamp := "1700000001"
	nonce := "nonce"
	sig := signature(token, timestamp, nonce, encrypted)

	body := []byte("<xml>" +
		"<ToUserName><![CDATA[to]]></ToUserName>" +
		"<Encrypt><![CDATA[" + encrypted + "]]></Encrypt>" +
		"</xml>")

	req := httptest.NewRequest(http.MethodPost, "/wecom/callback?msg_signature="+sig+"&timestamp="+timestamp+"&nonce="+nonce, bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if got := core.Calls(); got != 1 {
		t.Fatalf("core calls = %d, want 1", got)
	}
}

func TestCallbackHandler_ConcurrentDedupByMsgID(t *testing.T) {
	t.Parallel()

	token := "test-token"
	crypto := mustTestCrypto(t, token, "ww123")
	core := &testCoreHandler{}
	deduper := NewDeduper(10 * time.Minute)
	t.Cleanup(deduper.Close)

	h := NewCallbackHandler(CallbackDeps{
		Crypto:  crypto,
		Core:    core,
		Deduper: deduper,
	})

	plain := []byte("<xml>" +
		"<ToUserName><![CDATA[to]]></ToUserName>" +
		"<FromUserName><![CDATA[user]]></FromUserName>" +
		"<CreateTime>1700000000</CreateTime>" +
		"<MsgType><![CDATA[text]]></MsgType>" +
		"<Content><![CDATA[菜单]]></Content>" +
		"<MsgId>12345</MsgId>" +
		"</xml>")
	encrypted := mustEncrypt(t, crypto, plain)

	timestamp := "1700000001"
	nonce := "nonce"
	sig := signature(token, timestamp, nonce, encrypted)

	body := []byte("<xml>" +
		"<ToUserName><![CDATA[to]]></ToUserName>" +
		"<Encrypt><![CDATA[" + encrypted + "]]></Encrypt>" +
		"</xml>")

	const n = 20
	var wg sync.WaitGroup
	wg.Add(n)
	errCh := make(chan int, n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodPost, "/wecom/callback?msg_signature="+sig+"&timestamp="+timestamp+"&nonce="+nonce, bytes.NewReader(body))
			w := httptest.NewRecorder()
			h.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				errCh <- w.Code
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for code := range errCh {
		t.Fatalf("status = %d, want %d", code, http.StatusOK)
	}

	if got := core.Calls(); got != 1 {
		t.Fatalf("core calls = %d, want 1", got)
	}
}

func mustTestCrypto(t *testing.T, token string, receiverID string) *Crypto {
	t.Helper()
	rawKey := []byte("0123456789abcdef0123456789abcdef")
	encodingAESKey := strings.TrimRight(base64.StdEncoding.EncodeToString(rawKey), "=")
	crypto, err := NewCrypto(CryptoConfig{
		Token:          token,
		EncodingAESKey: encodingAESKey,
		ReceiverID:     receiverID,
	})
	if err != nil {
		t.Fatalf("NewCrypto() error: %v", err)
	}
	return crypto
}

func mustEncrypt(t *testing.T, crypto *Crypto, plain []byte) string {
	t.Helper()
	encrypted, err := crypto.Encrypt(plain, []byte("0123456789abcdef"))
	if err != nil {
		t.Fatalf("Encrypt() error: %v", err)
	}
	return encrypted
}
