package user

import (
	"context"
	"net/http"

	"github.com/thoas/go-funk"

	"github.com/swagftw/covax19-blockchain/pkg/wallet"
	"github.com/swagftw/covax19-blockchain/types"
	"github.com/swagftw/covax19-blockchain/utl/server/fault"
	"github.com/swagftw/covax19-blockchain/utl/transaction"
)

var (
	errUserNotFound = "ERR_USER_NOT_FOUND"
)

// service represents user service struct.
type service struct {
	tx   transaction.Transaction
	repo Repository
}

func (s service) GetUsersByAddresses(ctx context.Context, addresses []string) ([]*types.User, error) {
	users, err := s.repo.GetUsersByAddresses(ctx, addresses)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (s service) GetUserByWallet(ctx context.Context, wallet string) (*types.User, error) {
	user, err := s.repo.GetUserByWallet(ctx, wallet)
	if err != nil {
		if err == types.ErrUserNotFound {
			return nil, fault.New(errUserNotFound, "user not found", http.StatusNotFound)
		}

		return nil, err
	}

	return user, nil
}

// CheckPassword checks if user exists and password is correct.
func (s service) CheckPassword(ctx context.Context, userID uint, password string) (bool, error) {
	pass, err := s.repo.GetUserPassword(ctx, userID)
	if err != nil {
		return false, err
	}

	if pass != password {
		return false, nil
	}

	return true, nil
}

func (s service) CreateUser(ctx context.Context, user *types.CreateUserRequestDto) (*types.User, error) {
	// depending on type of user, create a new user.
	switch user.Type {
	case types.UserTypeCitizen:
		usr, err := s.createCitizen(ctx, user)

		return usr, err
	case types.UserTypeManufacturer:
		usr, err := s.createManufacturer(ctx, user)

		return usr, err
	case types.UserTypeMedicalInstitution:
		usr, err := s.createMedicalInstitution(ctx, user)

		return usr, err
	case types.UserTypeGovernment:
		return nil, types.ErrInvalidUserType
	default:
		return nil, types.ErrInvalidUserType
	}
}

// GetUserByEmail gets user by email.
func (s service) GetUserByEmail(ctx context.Context, email string) (*types.User, error) {
	usr, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if err == types.ErrUserNotFound {
			return nil, types.ErrUserNotFound
		}

		return nil, err
	}

	return usr, nil
}

// createManufacturer creates a new manufacturer.
func (s service) createManufacturer(ctx context.Context, dto *types.CreateUserRequestDto) (*types.User, error) {
	dto.Type = types.UserTypeManufacturer
	dto.WalletAddress = wallet.GenerateNewWallet()

	user, err := s.repo.CreateUser(ctx, dto)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// createMedicalInstitution creates new medical institution.
func (s service) createMedicalInstitution(ctx context.Context, dto *types.CreateUserRequestDto) (*types.User, error) {
	dto.Type = types.UserTypeMedicalInstitution
	dto.WalletAddress = wallet.GenerateNewWallet()

	user, err := s.repo.CreateUser(ctx, dto)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// createCitizen creates new citizen.
func (s service) createCitizen(ctx context.Context, dto *types.CreateUserRequestDto) (*types.User, error) {
	dto.Type = types.UserTypeCitizen
	dto.Verified = true
	dto.WalletAddress = wallet.GenerateNewWallet()

	user, err := s.repo.CreateUser(ctx, dto)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUser returns user by id.
func (s service) GetUser(ctx context.Context, id uint) (*types.User, error) {
	user, err := s.repo.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUsers returns all user.
func (s service) GetUsers(ctx context.Context, userType string) ([]*types.User, error) {
	if !funk.Contains(types.ValidUserTypes, types.UserType(userType)) {
		return nil, types.ErrInvalidUserType
	}

	users, err := s.repo.GetUsersByType(ctx, userType)
	if err != nil {
		return nil, err
	}

	return users, nil
}

// UpdateUser updates user.
func (s service) UpdateUser(ctx context.Context, user *types.User) (*types.User, error) {
	// TODO implement me
	panic("implement me")
}

// NewService returns new user service.
func NewService(tx transaction.Transaction, repo Repository) types.UserService {
	return &service{
		tx:   tx,
		repo: repo,
	}
}
