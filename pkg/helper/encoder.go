package helper

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
)

func EncryptID(plainText string) (string, error) {
	block, err := aes.NewCipher(SECRET_KEY)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, 12) // Nonce 12 bytes cho GCM
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	cipherText := aesGCM.Seal(nil, nonce, []byte(plainText), nil)
	result := append(nonce, cipherText...)

	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(result), nil
}
