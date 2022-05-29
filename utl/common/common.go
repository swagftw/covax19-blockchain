package common

import (
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

func MapJwtClaimToEchoContext(c echo.Context, token *jwt.Token) echo.Context {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c
	}

	id, ok := claims["id"].(string)
	if ok {
		c.Set("id", id)
	}

	username, ok := claims["name"].(string)
	if ok {
		c.Set("username", username)
	}

	email, ok := claims["email"].(string)
	if ok {
		c.Set("email", email)
	}

	aadhaar, ok := claims["aadhaar"].(string)
	if ok {
		c.Set("aadhaar", aadhaar)
	}

	wallet, ok := claims["wallet"].(string)
	if ok {
		c.Set("wallet", wallet)
	}

	userType, ok := claims["type"].(string)
	if ok {
		c.Set("type", userType)
	}

	return c
}
