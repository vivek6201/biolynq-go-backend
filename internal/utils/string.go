package utils

import (
	"crypto/rand"
	"math/big"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// GenerateRandomString generates a cryptographically secure random alphanumeric string of length n.
func GenerateRandomString(n int) (string, error) {
	sb := make([]byte, n)
	for i := range n {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		sb[i] = charset[num.Int64()]
	}
	return string(sb), nil
}
