package handlers

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
