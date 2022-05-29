package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/swagftw/covax19-blockchain/pkg/auth"
	"github.com/swagftw/covax19-blockchain/utl/storage"
)

type repo struct {
	db *gorm.DB
}

func (r repo) GetTokenByIdentifier(ctx context.Context, identifier string) (*auth.Tokens, error) {
	db := storage.GetGormDBFromContext(ctx, r.db)

	var t auth.Tokens
	if err := db.Where("identifier = ?", identifier).First(&t).Error; err != nil {
		return nil, err
	}

	return &t, nil
}

func (r repo) SaveTokens(ctx context.Context, tokens *auth.Tokens) error {
	db := storage.GetGormDBFromContext(ctx, r.db)

	return db.Save(tokens).Error
}

func NewRepo(db *gorm.DB) auth.Repository {
	return &repo{db: db}
}
