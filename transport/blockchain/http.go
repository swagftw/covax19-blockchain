package blockchain

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/swagftw/covax19-blockchain/pkg/blockchain/network"
	"github.com/swagftw/covax19-blockchain/types"
	"github.com/swagftw/covax19-blockchain/utl/server"
)

type httpHandler struct {
	usrService types.UserService
}

// NewHTTP initializes all the handlers.
func NewHTTP(v1Group *echo.Group, userService types.UserService, jwtMiddleware echo.MiddlewareFunc) {
	h := &httpHandler{usrService: userService}
	v1Group.GET("/ping", h.ping)

	// wallet related handlers
	chainGroup := v1Group.Group("/chain", jwtMiddleware)
	chainGroup.POST("/wallets", h.createWallet)
	chainGroup.GET("/wallets", h.getWallets)
	chainGroup.GET("/wallets/balance/:address", h.getBalance)

	// blockchain related handlers
	chainGroup.POST("/:address", h.createBlockchain)
	chainGroup.GET("", h.getBlockchain)
}

func (h *httpHandler) getBalance(ctx echo.Context) error {
	address := ctx.Param("address")
	if address == "" {
		return errors.New("address is required")
	}

	endpoint := fmt.Sprintf("http://%s/v1/chain/wallets/balance/%s", network.KnownNodes[0], address)

	resp, err := server.SendRequest(http.MethodGet, endpoint, nil)

	if err != nil {
		return ctx.JSON(http.StatusBadRequest, err)
	}

	return ctx.JSON(http.StatusOK, resp)
}

// ping is a simple health check endpoint.
func (h *httpHandler) ping(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "pong")
}

// createWallet creates wallet and returns its address.
func (h *httpHandler) createWallet(ctx echo.Context) error {
	endpoint := fmt.Sprintf("http://%s/v1/chain/wallets", network.KnownNodes[0])
	resp, err := server.SendRequest(http.MethodPost, endpoint, nil)

	if err != nil {
		return ctx.JSON(http.StatusBadRequest, err)
	}

	return ctx.JSON(http.StatusOK, resp)
}

// getAllWallets returns all wallets.
func (h *httpHandler) getWallets(ctx echo.Context) error {
	endpoint := fmt.Sprintf("http://%s/v1/chain/wallets", network.KnownNodes[0])
	resp, err := server.SendRequest(http.MethodGet, endpoint, nil)

	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, resp)
}

// createBlockchain creates a new blockchain.
func (h *httpHandler) createBlockchain(ctx echo.Context) error {
	address := ctx.Param("address")

	endpoint := fmt.Sprintf("http://%s/v1/chain", network.KnownNodes[0])
	resp, err := server.SendRequest(http.MethodPost, endpoint, types.CreateBlockchain{Address: address})

	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusCreated, resp)
}

// send creates a transaction.
func (h *httpHandler) send(ctx echo.Context) error {
	endpoint := fmt.Sprintf("http://%s/v1/transactions/send", network.KnownNodes[0])
	sendTokens := new(types.SendTokens)

	if err := ctx.Bind(sendTokens); err != nil {
		return err
	}

	// get sender by address
	user, err := h.usrService.GetUserByWallet(server.ToGoContext(ctx), sendTokens.From)
	if err != nil {
		return err
	}

	if user.Type == types.UserTypeGovernment {
		sendTokens.SkipBalanceCheck = true
	}

	resp, err := server.SendRequest(http.MethodPost, endpoint, sendTokens)

	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, resp)
}

// getBlockchain returns the blockchain.
func (h *httpHandler) getBlockchain(ctx echo.Context) error {
	endpoint := fmt.Sprintf("http://%s/v1/chain", network.KnownNodes[0])

	resp, err := server.SendRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, err)
	}

	return ctx.JSON(http.StatusOK, resp)
}
