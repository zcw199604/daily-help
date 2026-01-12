package wecom

import (
	"context"
	"encoding/xml"
	"io"
	"net/http"
)

type CallbackDeps struct {
	Crypto *Crypto
	Core   MessageHandler
}

type MessageHandler interface {
	HandleMessage(ctx context.Context, msg IncomingMessage) error
}

type encryptedEnvelope struct {
	ToUserName string `xml:"ToUserName"`
	Encrypt    string `xml:"Encrypt"`
}

func NewCallbackVerifyHandler(crypto *Crypto) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		msgSignature := r.URL.Query().Get("msg_signature")
		timestamp := r.URL.Query().Get("timestamp")
		nonce := r.URL.Query().Get("nonce")
		echostr := r.URL.Query().Get("echostr")

		if !crypto.VerifySignature(msgSignature, timestamp, nonce, echostr) {
			http.Error(w, "invalid signature", http.StatusForbidden)
			return
		}

		plain, err := crypto.Decrypt(echostr)
		if err != nil {
			http.Error(w, "decrypt failed", http.StatusForbidden)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(plain)
	})
}

func NewCallbackHandler(deps CallbackDeps) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		msgSignature := r.URL.Query().Get("msg_signature")
		timestamp := r.URL.Query().Get("timestamp")
		nonce := r.URL.Query().Get("nonce")

		b, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		var env encryptedEnvelope
		if err := xml.Unmarshal(b, &env); err != nil {
			http.Error(w, "bad xml", http.StatusBadRequest)
			return
		}
		if env.Encrypt == "" {
			http.Error(w, "missing encrypt", http.StatusBadRequest)
			return
		}

		if !deps.Crypto.VerifySignature(msgSignature, timestamp, nonce, env.Encrypt) {
			http.Error(w, "invalid signature", http.StatusForbidden)
			return
		}

		plain, err := deps.Crypto.Decrypt(env.Encrypt)
		if err != nil {
			http.Error(w, "decrypt failed", http.StatusForbidden)
			return
		}

		var msg IncomingMessage
		if err := xml.Unmarshal(plain, &msg); err != nil {
			http.Error(w, "bad xml", http.StatusBadRequest)
			return
		}

		_ = deps.Core.HandleMessage(r.Context(), msg)

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	})
}
