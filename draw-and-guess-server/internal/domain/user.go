package domain

import "time"

// User represents a user in the system
type User struct {
	ID          int64     `json:"id" db:"id"`
	Username    string    `json:"username" db:"username"`
	DisplayName string    `json:"display_name" db:"display_name"`
	Email       string    `json:"email" db:"email"`
	AvatarURL   string    `json:"avatar_url" db:"avatar_url"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
