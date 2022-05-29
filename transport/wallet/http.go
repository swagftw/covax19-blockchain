package wallet

import (
	"github.com/labstack/echo/v4"
	"github.com/swagftw/covax19-blockchain/types"
	"github.com/swagftw/covax19-blockchain/utl/server"
)

type http struct {
	walletService types.WalletService
}

func NewHTTP(v1Group echo.Group, walletService types.WalletService) {
	h := &http{
		walletService: walletService,
	}

	walletGroup := v1Group.Group("/wallets")

	walletGroup.GET("", h.getWallets)
	walletGroup.GET("/:userId", h.getWallet)
}

func (h http) getWallets(c echo.Context) error {
	wallets, err := h.walletService.GetWalletAddresses(server.ToGoContext(c))
	if err != nil {
		return err
	}
	return c.JSON(200, wallets)
}

func (h http) getWallet(c echo.Context) error {
	userId := c.Param("userId")
	wallet, err := h.walletService.GetWallet(server.ToGoContext(c), userId)
	if err != nil {
		return err
	}
	return c.JSON(200, wallet)
}
