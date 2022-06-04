package users

import (
	"net/http"
	"strconv"

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

	userGroup.GET("/type/:type", h.getUsers, authMiddleware)
	userGroup.GET("/:id", h.getUser, authMiddleware)
	userGroup.GET("/wallet/:address", h.getUserByWallet, authMiddleware)
}

func (h httpHandler) getUsers(c echo.Context) error {
	users, err := h.userService.GetUsers(server.ToGoContext(c), c.Param("type"))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, users)
}

func (h httpHandler) getUser(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	user, err := h.userService.GetUser(server.ToGoContext(c), uint(id))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, user)
}

func (h httpHandler) getUserByWallet(c echo.Context) error {
	address := c.Param("address")

	user, err := h.userService.GetUserByWallet(server.ToGoContext(c), address)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, user)
}
