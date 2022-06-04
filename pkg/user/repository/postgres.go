package repository

import (
	"context"
	"database/sql"
	"log"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"

	"github.com/swagftw/covax19-blockchain/pkg/user"
	"github.com/swagftw/covax19-blockchain/types"
	"github.com/swagftw/covax19-blockchain/utl/storage"
)

type repo struct {
	db *gorm.DB
}

func (r repo) GetUsersByAddresses(ctx context.Context, addresses []string) ([]*types.User, error) {
	var users []*types.User
	if len(addresses) == 0 {
		return users, nil
	}

	db := storage.GetGormDBFromContext(ctx, r.db)

	err := db.Transaction(func(tx *gorm.DB) error {
		usrs := make([]*user.User, 0)
		if err := tx.Where("wallet_address IN (?)", addresses).Find(&usrs).Error; err != nil {
			return err
		}

		err := copier.Copy(&users, &usrs)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return users, err
	}

	return users, nil
}

func (r repo) GetUsersByType(ctx context.Context, userType string) ([]*types.User, error) {
	db := storage.GetGormDBFromContext(ctx, r.db)

	var users []*types.User

	err := db.Transaction(func(tx *gorm.DB) error {
		usrs := make([]*user.User, 0)
		err := tx.Where("type = ?", userType).Find(&usrs).Error
		if err != nil {
			return err
		}

		err = copier.Copy(&users, &usrs)
		if err != nil {
			return types.ErrCopy
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return users, nil
}

func (r repo) GetUserByWallet(ctx context.Context, wallet string) (*types.User, error) {
	db := storage.GetGormDBFromContext(ctx, r.db)

	resp := new(types.User)

	err := db.Transaction(func(tx *gorm.DB) error {
		usr := new(user.User)

		err := db.Model(usr).Where("wallet_address = ?", wallet).First(usr).Error
		if err == gorm.ErrRecordNotFound {
			return types.ErrUserNotFound
		}

		err = copier.Copy(resp, usr)
		if err != nil {
			return types.ErrCopy
		}

		return nil
	})

	return resp, err
}

func (r repo) GetUserPassword(ctx context.Context, id uint) (string, error) {
	db := storage.GetGormDBFromContext(ctx, r.db)

	usr := new(user.User)
	if err := db.Where("id = ?", id).Preload("Password").First(usr).Error; err != nil {
		return "", err
	}

	return usr.Password.Password, nil
}

// GetUserByEmail returns a user by email.
func (r repo) GetUserByEmail(ctx context.Context, email string) (*types.User, error) {
	db := storage.GetGormDBFromContext(ctx, r.db)
	respUser := new(types.User)

	err := db.Transaction(func(tx *gorm.DB) error {
		usr := new(user.User)
		err := db.Where("email = ?", email).First(usr).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return types.ErrUserNotFound
			}

			return err
		}

		err = copier.Copy(respUser, usr)
		if err != nil {
			return types.ErrCopy
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return respUser, nil
}

// UpdateUser updates user in the database.
func (r repo) UpdateUser(ctx context.Context, dto *types.UpdateUserRequestDto) (*types.User, error) {
	db := storage.GetGormDBFromContext(ctx, r.db)
	usr := new(user.User)
	respUser := new(types.User)

	err := db.Transaction(func(tx *gorm.DB) error {
		err := copier.Copy(usr, dto)
		if err != nil {
			return types.ErrCopy
		}

		err = tx.Updates(usr).Error
		if err != nil {
			return types.ErrUpdatingUser
		}

		err = copier.Copy(respUser, usr)
		if err != nil {
			return types.ErrCopy
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return respUser, nil
}

// CreateUser creates a new user in the database.
func (r repo) CreateUser(ctx context.Context, dto *types.CreateUserRequestDto) (*types.User, error) {
	db := storage.GetGormDBFromContext(ctx, r.db)
	usr := new(user.User)
	respUser := new(types.User)

	err := db.Transaction(func(txn *gorm.DB) error {
		err := copier.Copy(usr, dto)
		if err != nil {
			return types.ErrCopy
		}

		usr.Password.Password = dto.Password

		if usr.AadhaarNumber.String == "" {
			usr.AadhaarNumber = sql.NullString{
				Valid: false,
			}
		}

		err = txn.Create(usr).Error
		if err != nil {
			log.Println(err)

			return types.ErrCreatingUser
		}

		err = copier.Copy(respUser, usr)
		if err != nil {
			return types.ErrCopy
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return respUser, nil
}

func (r repo) GetUser(ctx context.Context, id uint) (*types.User, error) {
	db := storage.GetGormDBFromContext(ctx, r.db)
	usr := new(user.User)

	err := db.Where("id = ?", id).First(usr).Error
	if err != nil {
		return nil, err
	}

	respUser := new(types.User)

	err = copier.Copy(respUser, usr)
	if err != nil {
		return nil, types.ErrCopy
	}

	return respUser, nil
}

func NewRepo(db *gorm.DB) user.Repository {
	return &repo{db: db}
}
