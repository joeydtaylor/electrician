package receivingrelay

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"fmt"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
)

func decryptData(data []byte, secOpts *relay.SecurityOptions, key string) ([]byte, error) {
	if secOpts == nil || !secOpts.Enabled || secOpts.Suite != ENCRYPTION_AES_GCM {
		return data, nil
	}
	aesKey := []byte(key)
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("decryptData: invalid AES key: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("decryptData: failed to create GCM: %w", err)
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("decryptData: ciphertext too short for nonce")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryptData: GCM decryption failed: %w", err)
	}
	return plaintext, nil
}
