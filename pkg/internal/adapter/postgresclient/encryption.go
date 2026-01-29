package postgresclient

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
	cseModeAESGCM = "aes-gcm"
)

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
