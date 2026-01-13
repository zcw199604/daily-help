package wecom

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"sort"
)

const pkcs7BlockSize = 32

type CryptoConfig struct {
	Token          string
	EncodingAESKey string
	ReceiverID     string
}

type Crypto struct {
	token      string
	aesKey     []byte
	receiverID string
}

func NewCrypto(cfg CryptoConfig) (*Crypto, error) {
	if cfg.Token == "" || cfg.EncodingAESKey == "" || cfg.ReceiverID == "" {
		return nil, errors.New("wecom crypto 配置不完整")
	}

	key, err := base64.StdEncoding.DecodeString(cfg.EncodingAESKey + "=")
	if err != nil {
		return nil, fmt.Errorf("decode encoding_aes_key: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("encoding_aes_key 解码后长度应为 32，实际为 %d", len(key))
	}

	return &Crypto{
		token:      cfg.Token,
		aesKey:     key,
		receiverID: cfg.ReceiverID,
	}, nil
}

func (c *Crypto) VerifySignature(msgSignature, timestamp, nonce, encrypted string) bool {
	expected := signature(c.token, timestamp, nonce, encrypted)
	return expected == msgSignature
}

func signature(token, timestamp, nonce, encrypted string) string {
	items := []string{token, timestamp, nonce, encrypted}
	sort.Strings(items)
	h := sha1.Sum([]byte(items[0] + items[1] + items[2] + items[3]))
	return fmt.Sprintf("%x", h)
}

func (c *Crypto) Decrypt(encryptedBase64 string) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedBase64)
	if err != nil {
		return nil, fmt.Errorf("base64 decode: %w", err)
	}

	block, err := aes.NewCipher(c.aesKey)
	if err != nil {
		return nil, err
	}

	if len(ciphertext)%block.BlockSize() != 0 {
		return nil, errors.New("ciphertext 非 blocksize 对齐")
	}

	iv := c.aesKey[:block.BlockSize()]
	mode := cipher.NewCBCDecrypter(block, iv)

	plain := make([]byte, len(ciphertext))
	mode.CryptBlocks(plain, ciphertext)

	plain, err = pkcs7Unpad(plain, pkcs7BlockSize)
	if err != nil {
		return nil, err
	}

	if len(plain) < 20 {
		return nil, errors.New("plaintext 长度不足")
	}

	msgLen := binary.BigEndian.Uint32(plain[16:20])
	if int(20+msgLen) > len(plain) {
		return nil, errors.New("msg_len 越界")
	}

	msg := plain[20 : 20+msgLen]
	recvID := plain[20+msgLen:]
	if string(recvID) != c.receiverID {
		return nil, errors.New("receiver_id 不匹配")
	}

	return msg, nil
}

func (c *Crypto) Encrypt(plaintext []byte, random16 []byte) (string, error) {
	if len(random16) != 16 {
		return "", errors.New("random16 必须为 16 字节")
	}

	msgLen := make([]byte, 4)
	binary.BigEndian.PutUint32(msgLen, uint32(len(plaintext)))

	payload := bytes.Join([][]byte{
		random16,
		msgLen,
		plaintext,
		[]byte(c.receiverID),
	}, nil)

	block, err := aes.NewCipher(c.aesKey)
	if err != nil {
		return "", err
	}

	payload = pkcs7Pad(payload, pkcs7BlockSize)

	iv := c.aesKey[:block.BlockSize()]
	mode := cipher.NewCBCEncrypter(block, iv)

	ciphertext := make([]byte, len(payload))
	mode.CryptBlocks(ciphertext, payload)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func pkcs7Pad(b []byte, blockSize int) []byte {
	padding := blockSize - (len(b) % blockSize)
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(b, padtext...)
}

func pkcs7Unpad(b []byte, blockSize int) ([]byte, error) {
	if len(b) == 0 || len(b)%blockSize != 0 {
		return nil, errors.New("invalid pkcs7 data")
	}
	padding := int(b[len(b)-1])
	if padding == 0 || padding > blockSize {
		return nil, errors.New("invalid pkcs7 padding")
	}
	for i := 0; i < padding; i++ {
		if b[len(b)-1-i] != byte(padding) {
			return nil, errors.New("invalid pkcs7 padding")
		}
	}
	return b[:len(b)-padding], nil
}
