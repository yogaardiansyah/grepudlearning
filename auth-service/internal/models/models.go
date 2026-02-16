package models

import "time"

type User struct {
	ID           int64     `json:"id"` //UUUID 32 <----
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	IsVerified   bool      `json:"is_verified"`
	CreatedAt    time.Time `json:"created_at"` 
}

type RefreshToken struct {
	ID                int64      `json:"id"`
	UserID            int64      `json:"user_id"`
	TokenHash         string     `json:"-"`
	DeviceID          string     `json:"device_id"`
	ExpiresAt         time.Time  `json:"expires_at"` 
	AbsoluteExpiresAt time.Time  `json:"absolute_expires_at"`
	RevokedAt         *time.Time `json:"revoked_at"`
}

