package jwt

import (
	"crypto/sha256"
	"fmt"
	"hash"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/golang-jwt/jwt"

	"github.com/swagftw/covax19-blockchain/types"
)

type Service struct {
	key    []byte
	algo   jwt.SigningMethod
	ttl    time.Duration
	hasher hash.Hash
}

// New generates new JWT service necessary for auth middleware.
func New() (Service, error) {
	signingMethod := jwt.GetSigningMethod("HS256")
	if signingMethod == nil {
		return Service{}, errors.New("invalid jwt signing method " + "HS256")
	}

	return Service{
		key:    []byte("abcd1234"),
		algo:   signingMethod,
		ttl:    time.Duration(15) * time.Minute,
		hasher: sha256.New(),
	}, nil
}

// ParseToken parses token from Authorization header.
func (s Service) ParseToken(authHeader string) (*jwt.Token, error) {
	parts := strings.SplitN(authHeader, " ", 2)
	if !(len(parts) == 2 && parts[0] == "Bearer") {
		return nil, errors.New("invalid authorization header")
	}

	token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
		if s.algo != token.Method {
			return nil, errors.New("invalid token")
		}

		return s.key, nil
	})
	if err != nil {
		return nil, err
	}

	return token, nil
}

// GenerateAccessToken generates new JWT token and populates it with user data.
func (s Service) GenerateAccessToken(usr *types.User) (string, error) {
	tokenString, err := jwt.NewWithClaims(s.algo, jwt.MapClaims{
		"id":      usr.ID,
		"name":    usr.Name,
		"email":   usr.Email,
		"aadhaar": usr.AadhaarNumber,
		"type":    usr.Type,
		"wallet":  usr.WalletAddress,
		"exp":     time.Now().Add(s.ttl).Unix(),
	}).SignedString(s.key)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GenerateRefreshToken generates new unique token.
func (s Service) GenerateRefreshToken(str string) string {
	s.hasher.Reset()
	_, _ = fmt.Fprintf(s.hasher, "%s%s", str, strconv.Itoa(time.Now().Nanosecond()))

	return fmt.Sprintf("%x", s.hasher.Sum(nil))
}
