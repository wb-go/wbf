// Package helpers предоставляет вспомогательные функции общего назначения.
package helpers

import "github.com/google/uuid"

// CreateUUID создает новый случайный UUID.
func CreateUUID() string {
	return uuid.New().String()
}

// ParseUUID проверяет, является ли строка валидным UUID.
func ParseUUID(s string) error {
	_, err := uuid.Parse(s)
	return err
}
