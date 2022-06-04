package transaction

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/swagftw/covax19-blockchain/types"
	"github.com/swagftw/covax19-blockchain/utl/server"
)

type httpHandler struct {
	service    types.Service
	usrService types.UserService
}

func NewHTTP(v1Group *echo.Group, service types.Service, userService types.UserService, jwtMiddleware echo.MiddlewareFunc) {
	h := &httpHandler{service: service, usrService: userService}

	// transaction related handlers
	transactionGroup := v1Group.Group("/transactions", jwtMiddleware)
	transactionGroup.POST("/send", h.send)
	transactionGroup.GET("/:address", h.getTransactions)
}

// send creates a transaction.
func (h *httpHandler) send(ctx echo.Context) error {
	sendTokens := new(types.SendTokens)

	if err := ctx.Bind(sendTokens); err != nil {
		return err
	}

	err := h.service.Send(server.ToGoContext(ctx), sendTokens)
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"message": "success!",
	})
}

func (h *httpHandler) getTransactions(c echo.Context) error {
	resp, err := h.service.GetTransaction(server.ToGoContext(c), c.Param("address"))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, resp)
}
