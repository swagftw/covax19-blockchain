package user

import (
	"context"
	"database/sql"

	"gorm.io/gorm"

	"github.com/swagftw/covax19-blockchain/types"
)

type (
	Repository interface {
		GetUser(ctx context.Context, id string) (*types.User, error)
		GetUserByEmail(ctx context.Context, email string) (*types.User, error)
		GetUserByWallet(ctx context.Context, wallet string) (*types.User, error)
		GetUserPassword(ctx context.Context, userID uint) (string, error)

		CreateUser(ctx context.Context, user *types.CreateUserRequestDto) (*types.User, error)

		UpdateUser(ctx context.Context, user *types.UpdateUserRequestDto) (*types.User, error)
	}

	User struct {
		ID            uint `gorm:"primaryKey"`
		Name          string
		Email         string         `gorm:"uniqueIndex"`
		Type          types.UserType `gorm:"default:'citizen';not null"`
		AadhaarNumber sql.NullString `gorm:"uniqueIndex"`
		WalletAddress string         `gorm:"uniqueIndex"`
		Verified      bool
		PasswordID    uint
		Password      *Password
		gorm.Model
	}

	Password struct {
		ID       uint   `gorm:"primaryKey"`
		Password string `gorm:"not null"`
		gorm.Model
	}
)

func (*User) TableName() string {
	return "usr.user"
}

func (*Password) TableName() string {
	return "usr.passwords"
}
