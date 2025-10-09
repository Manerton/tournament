package utils

import (
	"crypto/rand"
	"encoding/base64"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GenerateToken(length int) string {
	// Создаём буфер для случайных байт
	tokenBytes := make([]byte, length)

	// Заполняем случайными данными
	if _, err := rand.Read(tokenBytes); err != nil {
		return ""
	}

	// Кодируем в base64 (URL-safe, без '=' на конце)
	token := base64.RawURLEncoding.EncodeToString(tokenBytes)

	return token
}
