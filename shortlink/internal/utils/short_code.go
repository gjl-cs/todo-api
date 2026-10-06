package utils

import (
	"crypto/rand"
	"errors"
	"math/big"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateShortCode(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("short code length must be positive")
	}
	code := make([]byte, length)
	charsetLength := big.NewInt(int64(len(charset)))
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, charsetLength)
		if err != nil {
			return "", err
		}
		code[i] = charset[num.Int64()]
	}
	return string(code), nil
}
