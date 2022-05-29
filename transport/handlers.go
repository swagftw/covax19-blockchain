package transport

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/swagftw/covax19-blockchain/pkg/blockchain/network"
	"github.com/swagftw/covax19-blockchain/types"
	"github.com/swagftw/covax19-blockchain/utl/server"
)

// InitHandlers initializes all the handlers.
func InitHandlers(ech *echo.Echo) {
	ech.GET("/ping", ping)

	v1Group := ech.Group("/v1")

	// wallet related handlers
	chainGroup := v1Group.Group("/chain")
	chainGroup.POST("/wallets", createWallet)
	chainGroup.GET("/wallets", getWallets)
	chainGroup.GET("/wallets/balance/:address", getBalance)

	// blockchain related handlers
	chainGroup.POST("/:address", createBlockchain)
	chainGroup.GET("", getBlockchain)

	// transaction related handlers
	transactionGroup := v1Group.Group("/transactions")
	transactionGroup.POST("/send", send)
}

func getBalance(ctx echo.Context) error {
	address := ctx.Param("address")
	if address == "" {
		return errors.New("address is required")
	}

	endpoint := fmt.Sprintf("http://%s/v1/chain/wallets/balance/%s", network.KnownNodes[0], address)

	resp, err := server.SendRequest(http.MethodGet, endpoint, nil)

	if err == server.ErrBadRequest {
		return ctx.JSON(http.StatusBadRequest, err)
	}

	return ctx.JSON(http.StatusOK, resp)
}

// ping is a simple health check endpoint.
func ping(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "pong")
}

// createWallet creates wallet and returns its address.
func createWallet(ctx echo.Context) error {
	endpoint := fmt.Sprintf("http://%s/v1/chain/wallets", network.KnownNodes[0])
	resp, err := server.SendRequest(http.MethodPost, endpoint, nil)

	if err != server.ErrBadRequest {
		return ctx.JSON(http.StatusBadRequest, err)
	}

	return ctx.JSON(http.StatusOK, resp)
}

// getAllWallets returns all wallets.
func getWallets(ctx echo.Context) error {
	endpoint := fmt.Sprintf("http://%s/v1/chain/wallets", network.KnownNodes[0])
	resp, err := server.SendRequest(http.MethodGet, endpoint, nil)

	if err != server.ErrBadRequest {
		return err
	}

	return ctx.JSON(http.StatusOK, resp)
}

// createBlockchain creates a new blockchain.
func createBlockchain(ctx echo.Context) error {
	address := ctx.Param("address")

	endpoint := fmt.Sprintf("http://%s/v1/chain", network.KnownNodes[0])
	resp, err := server.SendRequest(http.MethodPost, endpoint, types.CreateBlockchain{Address: address})

	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusCreated, resp)
}

// send creates a transaction.
func send(ctx echo.Context) error {
	endpoint := fmt.Sprintf("http://%s/v1/transactions/send", network.KnownNodes[0])
	sendTokens := new(types.SendTokens)

	if err := ctx.Bind(sendTokens); err != nil {
		return err
	}

	resp, err := server.SendRequest(http.MethodPost, endpoint, sendTokens)

	if err != server.ErrBadRequest {
		return ctx.JSON(http.StatusBadRequest, err)
	}

	return ctx.JSON(http.StatusOK, resp)
}

// getBlockchain returns the blockchain.
func getBlockchain(ctx echo.Context) error {
	node := ctx.QueryParam("nodeID")
	if node == "" {
		return errors.New("nodeID is required")
	}

	endpoint := fmt.Sprintf("http://localhost:%s/v1/chain", node)

	resp, err := server.SendRequest(http.MethodGet, endpoint, nil)
	if err != server.ErrBadRequest {
		return ctx.JSON(http.StatusBadRequest, err)
	}

	return ctx.JSON(http.StatusOK, resp)
}
