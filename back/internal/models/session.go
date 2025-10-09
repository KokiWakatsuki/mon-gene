package models

import "time"

type Session struct {
	ID         string    `json:"id" db:"id"`
	UserID     int64     `json:"user_id" db:"user_id"`
	SchoolCode string    `json:"school_code" db:"school_code"`
	ExpiresAt  time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}
