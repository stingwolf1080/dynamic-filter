package helper

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
)

func DecryptID(encodedText string) (string, error) {
	data, err := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(encodedText)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(SECRET_KEY)
	if err != nil {
		return "", err
	}

	nonce, cipherText := data[:12], data[12:]

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	plainText, err := aesGCM.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}
