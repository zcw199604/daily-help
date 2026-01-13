package wecom

import (
	"context"
	"crypto/sha256"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

type CallbackDeps struct {
	Crypto  *Crypto
	Core    MessageHandler
	Deduper *Deduper
	// MaxBodyBytes 限制回调请求体大小，避免恶意超大 body 导致内存/CPU 被占满。默认 1MiB。
	MaxBodyBytes int64
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
		msgSignature := strings.TrimSpace(r.URL.Query().Get("msg_signature"))
		timestamp := strings.TrimSpace(r.URL.Query().Get("timestamp"))
		nonce := strings.TrimSpace(r.URL.Query().Get("nonce"))
		echostr := strings.TrimSpace(r.URL.Query().Get("echostr"))

		if !crypto.VerifySignature(msgSignature, timestamp, nonce, echostr) {
			slog.Warn("wecom callback verify 验签失败",
				"timestamp", timestamp,
				"nonce", nonce,
				"msg_signature_len", len(msgSignature),
				"echostr_len", len(echostr),
			)
			http.Error(w, "invalid signature", http.StatusForbidden)
			return
		}

		plain, err := crypto.Decrypt(echostr)
		if err != nil {
			slog.Warn("wecom callback verify 解密 echostr 失败",
				"error", err,
				"timestamp", timestamp,
				"nonce", nonce,
				"msg_signature_len", len(msgSignature),
				"echostr_len", len(echostr),
			)
			http.Error(w, "decrypt failed", http.StatusForbidden)
			return
		}

		slog.Info("wecom callback verify 成功",
			"timestamp", timestamp,
			"nonce", nonce,
			"msg_signature_len", len(msgSignature),
		)

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(plain)
	})
}

func NewCallbackHandler(deps CallbackDeps) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		msgSignature := strings.TrimSpace(r.URL.Query().Get("msg_signature"))
		timestamp := strings.TrimSpace(r.URL.Query().Get("timestamp"))
		nonce := strings.TrimSpace(r.URL.Query().Get("nonce"))

		maxBody := deps.MaxBodyBytes
		if maxBody <= 0 {
			maxBody = 1 << 20
		}
		r.Body = http.MaxBytesReader(w, r.Body, maxBody)

		b, err := io.ReadAll(r.Body)
		if err != nil {
			var maxErr *http.MaxBytesError
			if errors.As(err, &maxErr) {
				slog.Warn("wecom callback 读取请求体失败：payload too large",
					"error", err,
					"max_body_bytes", maxBody,
				)
				http.Error(w, "payload too large", http.StatusRequestEntityTooLarge)
				return
			}
			slog.Warn("wecom callback 读取请求体失败",
				"error", err,
				"max_body_bytes", maxBody,
			)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		var env encryptedEnvelope
		if err := xml.Unmarshal(b, &env); err != nil {
			slog.Warn("wecom callback 解析 envelope 失败（bad xml）",
				"error", err,
				"body_len", len(b),
			)
			http.Error(w, "bad xml", http.StatusBadRequest)
			return
		}
		env.Encrypt = strings.TrimSpace(env.Encrypt)
		if env.Encrypt == "" {
			slog.Warn("wecom callback 缺少 Encrypt 字段",
				"body_len", len(b),
			)
			http.Error(w, "missing encrypt", http.StatusBadRequest)
			return
		}

		if !deps.Crypto.VerifySignature(msgSignature, timestamp, nonce, env.Encrypt) {
			slog.Warn("wecom callback 验签失败",
				"timestamp", timestamp,
				"nonce", nonce,
				"msg_signature_len", len(msgSignature),
				"encrypt_len", len(env.Encrypt),
			)
			http.Error(w, "invalid signature", http.StatusForbidden)
			return
		}

		plain, err := deps.Crypto.Decrypt(env.Encrypt)
		if err != nil {
			slog.Warn("wecom callback 解密失败",
				"error", err,
				"timestamp", timestamp,
				"nonce", nonce,
				"msg_signature_len", len(msgSignature),
				"encrypt_len", len(env.Encrypt),
			)
			http.Error(w, "decrypt failed", http.StatusForbidden)
			return
		}

		var msg IncomingMessage
		if err := xml.Unmarshal(plain, &msg); err != nil {
			slog.Warn("wecom callback 解析明文消息失败（bad xml）",
				"error", err,
				"plain_len", len(plain),
			)
			http.Error(w, "bad xml", http.StatusBadRequest)
			return
		}

		content := strings.TrimSpace(msg.Content)
		normalized := strings.ToLower(content)
		isSelfTest := normalized == "ping" || normalized == "/ping" || normalized == "自检"

		slog.Info("wecom callback 已解析消息",
			"user_id", strings.TrimSpace(msg.FromUserName),
			"msg_type", strings.TrimSpace(msg.MsgType),
			"event", strings.TrimSpace(msg.Event),
			"event_key", strings.TrimSpace(msg.EventKey),
			"task_id", strings.TrimSpace(msg.TaskId),
			"response_code_len", len(strings.TrimSpace(msg.ResponseCode)),
			"msg_id", strings.TrimSpace(msg.MsgID),
			"content_len", len(content),
			"is_selftest", isSelfTest,
		)

		key := callbackDedupeKey(msg, plain)
		if deps.Deduper != nil && key != "" {
			if deps.Deduper.SeenOrMark(key) {
				slog.Info("wecom callback 重复消息已忽略",
					"user_id", strings.TrimSpace(msg.FromUserName),
					"msg_type", strings.TrimSpace(msg.MsgType),
					"event", strings.TrimSpace(msg.Event),
					"event_key", strings.TrimSpace(msg.EventKey),
					"task_id", strings.TrimSpace(msg.TaskId),
					"response_code", strings.TrimSpace(msg.ResponseCode),
					"msg_id", strings.TrimSpace(msg.MsgID),
				)
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("success"))
				return
			}
		}

		if err := deps.Core.HandleMessage(r.Context(), msg); err != nil {
			slog.Error("wecom callback 处理失败（不触发重试）",
				"error", err,
				"user_id", strings.TrimSpace(msg.FromUserName),
				"msg_type", strings.TrimSpace(msg.MsgType),
				"event", strings.TrimSpace(msg.Event),
				"event_key", strings.TrimSpace(msg.EventKey),
				"task_id", strings.TrimSpace(msg.TaskId),
				"response_code", strings.TrimSpace(msg.ResponseCode),
				"msg_id", strings.TrimSpace(msg.MsgID),
			)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	})
}

func callbackDedupeKey(msg IncomingMessage, plain []byte) string {
	user := strings.TrimSpace(msg.FromUserName)
	if tid := strings.TrimSpace(msg.TaskId); tid != "" {
		return fmt.Sprintf("task:%s:%s", user, tid)
	}
	if mid := strings.TrimSpace(msg.MsgID); mid != "" {
		return fmt.Sprintf("msg:%s:%s", user, mid)
	}
	if len(plain) > 0 {
		sum := sha256.Sum256(plain)
		return fmt.Sprintf("sha256:%x", sum[:])
	}
	return ""
}
