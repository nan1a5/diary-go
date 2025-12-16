package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

// Encrypt AES-256-GCM 加密
// key: 32 bytes
// plaintext: 明文
// 返回: ciphertext, nonce(IV), error
func Encrypt(key []byte, plaintext []byte) ([]byte, []byte, error) {
	if len(key) != 32 {
		return nil, nil, errors.New("key length must be 32 bytes")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce, nil
}

// Decrypt AES-256-GCM 解密
// key: 32 bytes
// ciphertext: 密文
// nonce: IV
// 返回: plaintext, error
func Decrypt(key []byte, ciphertext []byte, nonce []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, errors.New("key length must be 32 bytes")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(nonce) != gcm.NonceSize() {
		return nil, errors.New("invalid nonce length")
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
