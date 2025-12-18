package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
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

// EncryptToString 加密字符串并返回 Base64 编码的 (Nonce + Ciphertext)
func EncryptToString(key []byte, plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	encrypted, nonce, err := Encrypt(key, []byte(plaintext))
	if err != nil {
		return "", err
	}
	// Combine nonce + encrypted
	combined := append(nonce, encrypted...)
	return base64.StdEncoding.EncodeToString(combined), nil
}

// DecryptFromString 解密 Base64 编码的 (Nonce + Ciphertext)
func DecryptFromString(key []byte, cryptoText string) (string, error) {
	if cryptoText == "" {
		return "", nil
	}
	data, err := base64.StdEncoding.DecodeString(cryptoText)
	if err != nil {
		return "", err
	}
	// GCM nonce size is usually 12 bytes
	if len(data) < 12 {
		return "", errors.New("invalid ciphertext length")
	}
	nonce := data[:12]
	ciphertext := data[12:]
	plaintext, err := Decrypt(key, ciphertext, nonce)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
