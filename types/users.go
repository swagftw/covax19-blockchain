package types

import (
	"context"
	"errors"
)

var ErrCopy = errors.New("error copying struct")
var ErrUserNotFound = errors.New("user not found")
var ErrCreatingUser = errors.New("error creating user")
var ErrUpdatingUser = errors.New("error updating user")
var ErrInvalidUserType = errors.New("invalid user type")
var ErrInvalidPassword = errors.New("invalid password")
var ErrUserAlreadyExists = errors.New("user already exists")

type (
	// UserService is an interface for interacting with the user service.
	UserService interface {
		GetUser(ctx context.Context, id uint) (*User, error)
		GetUsers(ctx context.Context, userType string) ([]*User, error)
		GetUsersByAddresses(ctx context.Context, addresses []string) ([]*User, error)
		GetUserByEmail(ctx context.Context, email string) (*User, error)
		GetUserByWallet(ctx context.Context, wallet string) (*User, error)
		CreateUser(ctx context.Context, user *CreateUserRequestDto) (*User, error)
		UpdateUser(ctx context.Context, user *User) (*User, error)
		CheckPassword(ctx context.Context, userID uint, password string) (bool, error)
	}

	// CreateUserRequestDto dto for creating user.
	CreateUserRequestDto struct {
		Name          string   `json:"name,omitempty"`
		Email         string   `json:"email,omitempty"`
		AadhaarNumber string   `json:"aadhaarNumber,omitempty"`
		WalletAddress string   `json:"walletAddress,omitempty"`
		Password      string   `json:"password,omitempty"`
		Type          UserType `json:"type,omitempty"`
		Verified      bool     `json:"verified"`
	}

	UpdateUserRequestDto struct {
		ID       string `json:"id"`
		Verified bool   `json:"verified"`
	}

	// User represents a user dto.
	User struct {
		ID            uint     `json:"id,omitempty"`
		Name          string   `json:"name,omitempty"`
		Email         string   `json:"email,omitempty"`
		Type          UserType `json:"type,omitempty"`
		AadhaarNumber string   `json:"aadhaarNumber,omitempty"`
		WalletAddress string   `json:"walletAddress,omitempty"`
		Verified      bool     `json:"verified"`
	}

	// UserType represents a user type.
	UserType string
)

const (
	UserTypeGovernment         UserType = "government"
	UserTypeManufacturer       UserType = "manufacturer"
	UserTypeMedicalInstitution UserType = "medical_institution"
	UserTypeCitizen            UserType = "citizen"
)

var ValidUserTypes = []UserType{UserTypeManufacturer, UserTypeGovernment, UserTypeMedicalInstitution, UserTypeCitizen}
