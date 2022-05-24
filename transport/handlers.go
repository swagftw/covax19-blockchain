package transport

import (
	errWrap "errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/swagftw/covax19-blockchain/network"
	"github.com/swagftw/covax19-blockchain/types"
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

	resp, err := sendRequest(http.MethodGet, endpoint, nil)

	if errWrap.Is(err, ErrBadRequest) {
		return errors.Wrap(ctx.JSON(http.StatusBadRequest, err), "failed to get balance")
	}

	return errors.Wrap(ctx.JSON(http.StatusOK, resp), "failed to get balance")
}

// ping is a simple health check endpoint.
func ping(ctx echo.Context) error {
	return errors.Wrap(ctx.String(http.StatusOK, "pong"), "error in ping")
}

// createWallet creates wallet and returns its address.
func createWallet(ctx echo.Context) error {
	endpoint := fmt.Sprintf("http://%s/v1/chain/wallets", network.KnownNodes[0])
	resp, err := sendRequest(http.MethodPost, endpoint, nil)

	if errWrap.Is(err, ErrBadRequest) {
		return errors.Wrap(ctx.JSON(http.StatusBadRequest, err), "error in createWallet")
	}

	return errors.Wrap(ctx.JSON(http.StatusOK, resp), "error in createWallet")
}

// getAllWallets returns all wallets.
func getWallets(ctx echo.Context) error {
	endpoint := fmt.Sprintf("http://%s/v1/chain/wallets", network.KnownNodes[0])
	resp, err := sendRequest(http.MethodGet, endpoint, nil)

	if errWrap.Is(err, ErrBadRequest) {
		return errors.Wrap(ctx.JSON(http.StatusBadRequest, err), "error in getWallets")
	}

	return errors.Wrap(ctx.JSON(http.StatusOK, resp), "error in getWallets")
}

// createBlockchain creates a new blockchain.
func createBlockchain(ctx echo.Context) error {
	address := ctx.Param("address")

	endpoint := fmt.Sprintf("http://%s/v1/chain", network.KnownNodes[0])
	resp, err := sendRequest(http.MethodPost, endpoint, types.CreateBlockchain{Address: address})

	if err != nil {
		return errors.Wrap(ctx.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		}), "error in createBlockchain")
	}

	return errors.Wrap(ctx.JSON(http.StatusCreated, resp), "error in createBlockchain")
}

// send creates a transaction.
func send(ctx echo.Context) error {
	endpoint := fmt.Sprintf("http://%s/v1/transactions/send", network.KnownNodes[0])
	sendTokens := new(types.SendTokens)

	if err := ctx.Bind(sendTokens); err != nil {
		return errors.Wrap(ctx.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		}), "error in send")
	}

	resp, err := sendRequest(http.MethodPost, endpoint, sendTokens)

	if errWrap.Is(err, ErrBadRequest) {
		return errors.Wrap(ctx.JSON(http.StatusBadRequest, err), "error sending tokens")
	}

	return errors.Wrap(ctx.JSON(http.StatusOK, resp), "error sending tokens")
}

// getBlockchain returns the blockchain.
func getBlockchain(ctx echo.Context) error {
	node := ctx.QueryParam("nodeID")
	if node == "" {
		return errors.New("nodeID is required")
	}
	endpoint := fmt.Sprintf("http://localhost:%s/v1/chain", node)
	resp, err := sendRequest(http.MethodGet, endpoint, nil)

	if errWrap.Is(err, ErrBadRequest) {
		return errors.Wrap(ctx.JSON(http.StatusBadRequest, err), "error in getBlockchain")
	}

	return errors.Wrap(ctx.JSON(http.StatusOK, resp), "error in getBlockchain")
}
