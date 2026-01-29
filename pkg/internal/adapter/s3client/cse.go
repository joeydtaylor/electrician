package s3client

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
)

const (
	cseModeAESGCM          = "aes-gcm"
	cseMetaKey             = "x-electrician-cse"
	cseMetaContentType     = "x-electrician-content-type"
	cseMetaContentEncoding = "x-electrician-content-encoding"
)

func normalizeMeta(m map[string]string) map[string]string {
	if len(m) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[strings.ToLower(strings.TrimSpace(k))] = v
	}
	return out
}

func parseAESGCMKeyHex(keyHex string) ([]byte, error) {
	keyHex = strings.TrimSpace(keyHex)
	if keyHex == "" {
		return nil, fmt.Errorf("client-side encryption key is required")
	}
	raw, err := hex.DecodeString(keyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid client-side key hex: %w", err)
	}
	if len(raw) != 32 {
		return nil, fmt.Errorf("client-side key must be 32 bytes (AES-256)")
	}
	return raw, nil
}

func encryptAESGCM(plaintext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
	out := make([]byte, 0, len(nonce)+len(ciphertext))
	out = append(out, nonce...)
	out = append(out, ciphertext...)
	return out, nil
}

func decryptAESGCM(ciphertext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	ns := gcm.NonceSize()
	if len(ciphertext) < ns+gcm.Overhead() {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce := ciphertext[:ns]
	payload := ciphertext[ns:]
	return gcm.Open(nil, nonce, payload, nil)
}

func (a *S3Client[T]) validateWriterSecurityConfig() error {
	if a.configErr != nil {
		return a.configErr
	}
	if a.requireSSE && a.sseMode == "" {
		return fmt.Errorf("s3client: server-side encryption (SSE) is required")
	}
	if a.requireCSE && (a.cseMode == "" || len(a.cseKey) == 0) {
		return fmt.Errorf("s3client: client-side encryption is required")
	}
	if a.cseMode != "" && len(a.cseKey) == 0 {
		return fmt.Errorf("s3client: client-side encryption key is missing")
	}
	return nil
}

func (a *S3Client[T]) validateReaderSecurityConfig() error {
	if a.configErr != nil {
		return a.configErr
	}
	if a.requireCSE && (a.cseMode == "" || len(a.cseKey) == 0) {
		return fmt.Errorf("s3client: client-side encryption is required for reads")
	}
	return nil
}

func (a *S3Client[T]) applyCSE(payload []byte, contentType string, contentEncoding string) ([]byte, string, string, map[string]string, error) {
	if a.cseMode == "" {
		return payload, contentType, contentEncoding, nil, nil
	}
	switch strings.ToLower(a.cseMode) {
	case cseModeAESGCM:
		enc, err := encryptAESGCM(payload, a.cseKey)
		if err != nil {
			return nil, "", "", nil, err
		}
		meta := map[string]string{
			cseMetaKey: cseModeAESGCM,
		}
		if contentType != "" {
			meta[cseMetaContentType] = contentType
		}
		if contentEncoding != "" {
			meta[cseMetaContentEncoding] = contentEncoding
		}
		return enc, "application/octet-stream", "", meta, nil
	default:
		return nil, "", "", nil, fmt.Errorf("unsupported client-side encryption mode: %s", a.cseMode)
	}
}

func (a *S3Client[T]) decryptIfNeeded(meta map[string]string, payload []byte) ([]byte, string, string, error) {
	norm := normalizeMeta(meta)
	mode := strings.ToLower(strings.TrimSpace(norm[cseMetaKey]))
	if mode == "" {
		if a.requireCSE {
			return nil, "", "", fmt.Errorf("s3client: object missing client-side encryption metadata")
		}
		return payload, "", "", nil
	}
	if a.cseMode == "" || len(a.cseKey) == 0 {
		return nil, "", "", fmt.Errorf("s3client: client-side encryption key missing for encrypted object")
	}
	switch mode {
	case cseModeAESGCM:
		dec, err := decryptAESGCM(payload, a.cseKey)
		if err != nil {
			return nil, "", "", err
		}
		return dec, norm[cseMetaContentType], norm[cseMetaContentEncoding], nil
	default:
		return nil, "", "", fmt.Errorf("unsupported client-side encryption mode: %s", mode)
	}
}
