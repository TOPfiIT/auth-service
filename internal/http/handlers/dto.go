package handlers

import "time"

type RegisterRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type TokenRequest struct {
	AccessToken string `json:"access_token"`
}

type LoginRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type CreateRoomSessionRequest struct {
	RoomID   string    `json:"room_id"`
	ExpiryAt time.Time `json:"expiry_at"`
}
