package auth

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/swagftw/covax19-blockchain/types"
	"github.com/swagftw/covax19-blockchain/utl/server"
)

type httpHandler struct {
	authService types.AuthService
}

// NewHTTP initialize http handlers for auth service.
func NewHTTP(v1Group *echo.Group, authService types.AuthService) {
	h := &httpHandler{authService: authService}

	authGroup := v1Group.Group("/auth")

	authGroup.POST("/login", h.login)
	authGroup.POST("/signup", h.register)
}

func (h httpHandler) login(c echo.Context) error {
	req := new(types.LoginRequest)

	if err := c.Bind(req); err != nil {
		return err
	}

	resp, err := h.authService.Login(server.ToGoContext(c), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, resp)
}

func (h httpHandler) register(c echo.Context) error {
	req := new(types.SignUpRequest)

	if err := c.Bind(req); err != nil {
		return err
	}

	resp, err := h.authService.SignUp(server.ToGoContext(c), req)

	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, resp)
}
