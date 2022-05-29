package auth

import (
	"context"

	"gorm.io/gorm"
)

type (
	Repository interface {
		SaveTokens(ctx context.Context, tokens *Tokens) error
		GetTokenByIdentifier(ctx context.Context, identifier string) (*Tokens, error)
	}

	Tokens struct {
		ID           uint `gorm:"primary_key"`
		AccessToken  string
		RefreshToken string
		Identifier   string
		gorm.Model
	}
)

func (t *Tokens) TableName() string {
	return "auth.tokens"
}
