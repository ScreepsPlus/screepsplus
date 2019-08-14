package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

func pbkdf(password string, salt string) string {
	bytes := pbkdf2.Key([]byte(password), []byte(salt), 25000, 512, sha256.New)
	return hex.EncodeToString(bytes)
}

// HashPassword hashes password
func HashPassword(password string) (string, error) {
	saltBytes := make([]byte, 32)
	if _, err := rand.Read(saltBytes); err != nil {
		return "", err
	}
	salt := hex.EncodeToString(saltBytes)
	hash := pbkdf(password, salt)
	return fmt.Sprintf("%s.%s", salt, hash), nil
}

// VerifyPassword verifies password hash
func VerifyPassword(pass string, proposed string) (bool, error) {
	if !strings.ContainsRune(pass, '.') {
		return false, nil
	}
	parts := strings.Split(pass, ".")
	salt := parts[0]
	hash := parts[1]
	calcedHash := pbkdf(proposed, salt)
	valid := hash == calcedHash
	return valid, nil
}
