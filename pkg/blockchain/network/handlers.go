package network

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	blockchain2 "github.com/swagftw/covax19-blockchain/pkg/blockchain"
	wallet2 "github.com/swagftw/covax19-blockchain/pkg/wallet"
	"github.com/swagftw/covax19-blockchain/types"
	"github.com/swagftw/covax19-blockchain/utl/server/fault"
)

func (h HTTP) createWallet(c echo.Context) error {
	wallets, _ := wallet2.CreateWallets()
	wlt := wallets.AddWallet()
	wallets.SaveFile()

	return c.JSON(http.StatusCreated, map[string]string{
		"address": string(wlt.Address()),
	})
}

func (h HTTP) getWallets(c echo.Context) error {
	wallets, _ := wallet2.CreateWallets()
	addresses := wallets.GetAllAddresses()

	adrs := make([]string, 0)
	for _, address := range addresses {
		adrs = append(adrs, address)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"addresses": adrs,
	})
}

func (h HTTP) getBalance(c echo.Context) error {
	address := c.Param("address")
	if !wallet2.ValidateAddress(address) {
		log.Panic("Address is not Valid")
	}
	chain := h.chain
	UTXOSet := blockchain2.UTXOSet{Blockchain: chain}

	balance := 0
	pubKeyHash := wallet2.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := UTXOSet.FindUnspentTransactions(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"balance": balance,
	})
}

func (h HTTP) handleSend(c echo.Context) error {
	sendDTO := new(types.SendTokens)
	if err := c.Bind(sendDTO); err != nil {
		return err
	}

	if !wallet2.ValidateAddress(sendDTO.To) {
		log.Panic("Address is not Valid")
	}
	if !wallet2.ValidateAddress(sendDTO.From) {
		log.Panic("Address is not Valid")
	}
	chain := h.chain
	UTXOSet := blockchain2.UTXOSet{Blockchain: chain}

	wallets, err := wallet2.CreateWallets()
	if err != nil {
		log.Panic(err)
	}

	wallet := wallets.GetWallet(sendDTO.From)

	wallet2.DeleteWalletLock()

	tx, err := blockchain2.NewTransaction(wallet, sendDTO.To, sendDTO.Amount, &UTXOSet, sendDTO.SkipBalanceCheck)
	if err != nil {
		if err == types.ErrNotEnoughFunds {
			return fault.New("ERROR_NOT_ENOUGH_FUNDS", err.Error(), http.StatusBadRequest)
		}

		return err
	}

	// cbTx := blockchain2.CoinbaseTx(sendDTO.From, "")
	go func(UTXOSet blockchain2.UTXOSet, tx *blockchain2.Transaction) {
		memoryPool.mutex.Lock()
		txs := []*blockchain2.Transaction{tx}
		block := chain.MineBlock(txs)
		UTXOSet.Update(block)
		memoryPool.mutex.Unlock()
	}(UTXOSet, tx)

	log.Println("Success!")

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Success!",
	})
}

func (h HTTP) getChain(c echo.Context) error {
	chain := h.chain
	iter := chain.Iterator()

	resp := make([]*types.Block, 0)

	for {
		block := iter.Next()

		log.Printf("Hash: %x\n", block.Hash)
		log.Printf("Prev. hash: %x\n", block.PrevHash)
		pow := blockchain2.NewProof(block)
		log.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))

		for _, tx := range block.Transactions {
			log.Println(tx)
		}

		log.Println()

		resp = append(resp, &types.Block{
			PrevHash:  fmt.Sprintf("%x", block.PrevHash),
			Hash:      fmt.Sprintf("%x", block.Hash),
			PoW:       pow.Validate(),
			Timestamp: block.Timestamp,
		})

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"blocks": resp,
	})
}
