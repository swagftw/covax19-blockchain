package auth

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/swagftw/covax19-blockchain/types"
	"github.com/swagftw/covax19-blockchain/utl/jwt"
	"github.com/swagftw/covax19-blockchain/utl/server/fault"
	"github.com/swagftw/covax19-blockchain/utl/transaction"
)

var (
	ErrUserExists         = "ERR_USER_EXISTS"
	ErrInvalidCredentials = "ERR_INVALID_CREDENTIALS"
)

type service struct {
	tx          transaction.Transaction
	userService types.UserService
	repo        Repository
	jwt         jwt.Service
}

func (s service) SignUp(ctx context.Context, request *types.SignUpRequest) (*types.SignUpResponse, error) {
	resp := new(types.SignUpResponse)

	err := s.tx.Run(ctx, func(ctx context.Context) error {
		// check if user with email already exists.
		// err := s.tx.Run(ctx, func(ctx context.Context) error {
		usr, err := s.userService.GetUserByEmail(ctx, request.Email)
		// if yes return err
		if err != nil && err != types.ErrUserNotFound {
			return err
		} else if err == nil && usr != nil {
			return fault.New(ErrUserExists, "User with email already exists", http.StatusBadRequest)
		}

		// else create user and token.
		if usr == nil {
			dto := &types.CreateUserRequestDto{
				Name:          request.Name,
				Email:         request.Email,
				Password:      request.Password,
				AadhaarNumber: request.AadhaarNumber,
				Type:          types.UserType(request.Type),
			}

			usr, err = s.userService.CreateUser(ctx, dto)
			if err != nil {
				return err
			}
		}

		// create access token.
		token, err := s.jwt.GenerateAccessToken(usr)
		if err != nil {
			return err
		}

		// create refresh token.
		refreshToken := s.jwt.GenerateRefreshToken(token)

		// save tokens
		tokens := &Tokens{
			AccessToken:  token,
			RefreshToken: refreshToken,
			Identifier:   usr.Email,
		}

		err = s.repo.SaveTokens(ctx, tokens)
		if err != nil {
			return err
		}

		resp = &types.SignUpResponse{
			AccessToken:  token,
			RefreshToken: refreshToken,
			User:         usr,
		}

		httpError := new(fault.HTTPError)
		if errors.As(err, &httpError) {
			log.Println("error")
		}

		return nil
	})

	return resp, err
}

func (s service) Login(ctx context.Context, request *types.LoginRequest) (*types.LoginResponse, error) {
	resp := new(types.LoginResponse)

	// check if user with email already exists.
	err := s.tx.Run(ctx, func(ctx context.Context) error {
		usr, err := s.userService.GetUserByEmail(ctx, request.Email)
		if err != nil {
			return err
		}

		// check if password is correct.
		valid, err := s.userService.CheckPassword(ctx, usr.ID, request.Password)
		if err != nil {
			return err
		}

		if !valid {
			return fault.New(ErrInvalidCredentials, "invalid credentials", http.StatusBadRequest)
		}

		// create access token.
		token, err := s.jwt.GenerateAccessToken(usr)
		if err != nil {
			return err
		}

		// get tokens.
		tokens, err := s.repo.GetTokenByIdentifier(ctx, usr.Email)
		if err != nil {
			return err
		}

		// save token.
		tokens.AccessToken = token
		err = s.repo.SaveTokens(ctx, tokens)
		if err != nil {
			return err
		}

		resp = &types.LoginResponse{
			AccessToken:  token,
			RefreshToken: tokens.RefreshToken,
			User:         usr,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return resp, nil
}

func NewService(txn transaction.Transaction, repo Repository, userService types.UserService, jwtService jwt.Service) types.AuthService {
	return &service{
		tx:          txn,
		repo:        repo,
		jwt:         jwtService,
		userService: userService,
	}
}
