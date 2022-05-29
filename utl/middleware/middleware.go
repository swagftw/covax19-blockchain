package middleware

import (
	"errors"

	"github.com/labstack/echo/v4"

	"github.com/swagftw/covax19-blockchain/utl/common"
	"github.com/swagftw/covax19-blockchain/utl/jwt"
)

// JwtMiddleware makes JWT implement the JwtMiddleware interface.
func JwtMiddleware(tokenParser jwt.Service) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			tk := c.Request().Header.Get("Authorization")
			if tk == "" {
				tk = "Bearer " + c.QueryParam("token")
			}
			token, err := tokenParser.ParseToken(tk)
			if err != nil || !token.Valid {
				return errors.New("invalid token")
			}

			c = common.MapJwtClaimToEchoContext(c, token)

			return next(c)
		}
	}
}
