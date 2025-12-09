package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	"golang.org/x/crypto/scrypt"
)

const (
	passphraseEnvVar = "WIZQL_PASSPHRASE"
	saltSize         = 16
	nonceSize        = 12
)

func deriveKey(passphrase string, salt []byte) ([]byte, error) {
	if len(salt) == 0 {
		return nil, errors.New("missing salt for key derivation")
	}

	return scrypt.Key([]byte(passphrase), salt, 1<<15, 8, 1, 32)
}

func encryptSecret(plainText, passphrase string, salt []byte) (string, error) {
	key, err := deriveKey(passphrase, salt)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("unable to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("unable to create gcm: %w", err)
	}

	nonce := make([]byte, nonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("unable to generate nonce: %w", err)
	}

	sealed := gcm.Seal(nonce, nonce, []byte(plainText), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

func decryptSecret(cipherText, passphrase string, salt []byte) (string, error) {
	payload, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", fmt.Errorf("unable to decode secret: %w", err)
	}

	if len(payload) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce := payload[:nonceSize]
	encrypted := payload[nonceSize:]

	key, err := deriveKey(passphrase, salt)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("unable to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("unable to create gcm: %w", err)
	}

	decrypted, err := gcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return "", fmt.Errorf("unable to decrypt: %w", err)
	}

	return string(decrypted), nil
}

func generateSalt() ([]byte, error) {
	salt := make([]byte, saltSize)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("unable to generate salt: %w", err)
	}
	return salt, nil
}
