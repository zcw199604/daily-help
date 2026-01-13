package wecom

import (
	"crypto/sha1"
	"encoding/base64"
	"sort"
	"strings"
	"testing"
)

func TestCrypto_EncryptDecryptAndVerifySignature(t *testing.T) {
	t.Parallel()

	token := "test-token"
	receiverID := "ww1234567890"
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

	plaintext := []byte("<xml><MsgType><![CDATA[text]]></MsgType><Content><![CDATA[hello]]></Content></xml>")
	random16 := []byte("0123456789abcdef")

	encrypted, err := crypto.Encrypt(plaintext, random16)
	if err != nil {
		t.Fatalf("Encrypt() error: %v", err)
	}

	timestamp := "1700000000"
	nonce := "nonce"
	msgSignature := sha1Signature(token, timestamp, nonce, encrypted)

	if !crypto.VerifySignature(msgSignature, timestamp, nonce, encrypted) {
		t.Fatalf("VerifySignature() = false, want true")
	}

	got, err := crypto.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt() error: %v", err)
	}
	if string(got) != string(plaintext) {
		t.Fatalf("Decrypt() = %q, want %q", string(got), string(plaintext))
	}
}

func TestCrypto_PKCS7BlockSize32(t *testing.T) {
	t.Parallel()

	rawKey := []byte("0123456789abcdef0123456789abcdef")
	encodingAESKey := strings.TrimRight(base64.StdEncoding.EncodeToString(rawKey), "=")
	receiverID := "ww1234567890"

	crypto, err := NewCrypto(CryptoConfig{
		Token:          "t",
		EncodingAESKey: encodingAESKey,
		ReceiverID:     receiverID,
	})
	if err != nil {
		t.Fatalf("NewCrypto() error: %v", err)
	}

	plaintext := []byte("123456789012") // 12 bytes -> padding should be 20 (blockSize=32)
	encrypted, err := crypto.Encrypt(plaintext, []byte("0123456789abcdef"))
	if err != nil {
		t.Fatalf("Encrypt() error: %v", err)
	}

	cipherBytes, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		t.Fatalf("DecodeString() error: %v", err)
	}

	payloadLen := 16 + 4 + len(plaintext) + len(receiverID)
	padLen := 32 - (payloadLen % 32)
	expectedLen := payloadLen + padLen
	if len(cipherBytes) != expectedLen {
		t.Fatalf("ciphertext len = %d, want %d", len(cipherBytes), expectedLen)
	}
	if len(cipherBytes)%32 != 0 {
		t.Fatalf("ciphertext len %% 32 = %d, want 0", len(cipherBytes)%32)
	}

	got, err := crypto.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt() error: %v", err)
	}
	if string(got) != string(plaintext) {
		t.Fatalf("Decrypt() = %q, want %q", string(got), string(plaintext))
	}
}

func TestCrypto_ReceiverMismatch(t *testing.T) {
	t.Parallel()

	rawKey := []byte("0123456789abcdef0123456789abcdef")
	encodingAESKey := strings.TrimRight(base64.StdEncoding.EncodeToString(rawKey), "=")

	cryptoA, err := NewCrypto(CryptoConfig{
		Token:          "t",
		EncodingAESKey: encodingAESKey,
		ReceiverID:     "wwA",
	})
	if err != nil {
		t.Fatalf("NewCrypto(A) error: %v", err)
	}
	cryptoB, err := NewCrypto(CryptoConfig{
		Token:          "t",
		EncodingAESKey: encodingAESKey,
		ReceiverID:     "wwB",
	})
	if err != nil {
		t.Fatalf("NewCrypto(B) error: %v", err)
	}

	encrypted, err := cryptoA.Encrypt([]byte("<xml/>"), []byte("0123456789abcdef"))
	if err != nil {
		t.Fatalf("Encrypt() error: %v", err)
	}
	if _, err := cryptoB.Decrypt(encrypted); err == nil {
		t.Fatalf("Decrypt() error = nil, want mismatch error")
	}
}

func sha1Signature(token, timestamp, nonce, encrypted string) string {
	items := []string{token, timestamp, nonce, encrypted}
	sort.Strings(items)
	h := sha1.Sum([]byte(items[0] + items[1] + items[2] + items[3]))
	return strings.ToLower(base64Hex(h[:]))
}

func base64Hex(b []byte) string {
	const hexdigits = "0123456789abcdef"
	out := make([]byte, 0, len(b)*2)
	for _, v := range b {
		out = append(out, hexdigits[v>>4], hexdigits[v&0x0f])
	}
	return string(out)
}
