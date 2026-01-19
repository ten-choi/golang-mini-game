// Package utils provides trace ID generation for request tracking
package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// GenerateTraceID generates a unique trace ID for request tracking
// Format: timestamp-random (e.g., 20251225-a1b2c3d4)
func GenerateTraceID() string {
	timestamp := time.Now().Format("20060102150405")

	randomBytes := make([]byte, 4)
	if _, err := rand.Read(randomBytes); err != nil {
		// Fallback to timestamp only if random fails
		return timestamp
	}

	return fmt.Sprintf("%s-%s", timestamp, hex.EncodeToString(randomBytes))
}
