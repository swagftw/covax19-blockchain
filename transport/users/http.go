package users

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/swagftw/covax19-blockchain/types"
	"github.com/swagftw/covax19-blockchain/utl/server"
)

type httpHandler struct {
	userService types.UserService
}

func NewHTTP(v1Group *echo.Group, userService types.UserService, authMiddleware echo.MiddlewareFunc) {
	h := &httpHandler{
		userService: userService,
	}

	userGroup := v1Group.Group("/users")

	userGroup.GET("/:type", h.getUsers, authMiddleware)
	userGroup.GET("/:id", h.getUser, authMiddleware)
}

func (h httpHandler) getUsers(c echo.Context) error {
	users, err := h.userService.GetUsers(server.ToGoContext(c))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, users)
}

func (h httpHandler) getUser(c echo.Context) error {
	id := c.Param("id")

	user, err := h.userService.GetUser(server.ToGoContext(c), id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, user)
}
