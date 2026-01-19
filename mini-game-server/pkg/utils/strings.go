// Package utils provides common utility functions
package utils

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"unicode"
)

// GenerateRandomHexID generates a random hex ID
func GenerateRandomHexID(length int) (string, error) {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// IsKorean checks if a string contains only Korean characters
func IsKorean(s string) bool {
	for _, r := range s {
		if !unicode.Is(unicode.Hangul, r) && !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

// Sanitize removes special characters from a string
func Sanitize(s string) string {
	return strings.TrimSpace(s)
}

// Contains checks if a slice contains a string
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
