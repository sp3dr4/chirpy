package entities

import "time"

type RefreshToken struct {
	UserId    int       `json:"userId"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
}
