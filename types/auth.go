package types

import "context"

type (
	AuthService interface {
		SignUp(ctx context.Context, request *SignUpRequest) (*SignUpResponse, error)
		Login(ctx context.Context, request *LoginRequest) (*LoginResponse, error)
	}

	SignUpRequest struct {
		Email         string `json:"email"`
		Password      string `json:"password"`
		Name          string `json:"name"`
		AadhaarNumber string `json:"aadhaarNumber"`
		Type          string `json:"type"`
	}

	SignUpResponse struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
		User         *User  `json:"user"`
	}

	LoginResponse struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
		User         *User  `json:"user"`
	}

	LoginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
)
