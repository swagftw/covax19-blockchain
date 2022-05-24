package network

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors" //nolint:gci

	"github.com/swagftw/covax19-blockchain/blockchain"
	"github.com/swagftw/covax19-blockchain/types"
	"github.com/swagftw/covax19-blockchain/wallet"
)

// handleCmd handles inter-node commands and routes them.
//nolint:cyclop
func (h *HTTP) handleCmd(ctx echo.Context) error {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in HandleCmd", r)
			log.Println(string(debug.Stack()))
		}
	}()

	req := new(CmdRequest)
	if err := ctx.Bind(req); err != nil {
		return errors.Wrap(err, "failed to bind request")
	}

	log.Printf("Received command %s", req.Cmd)

	payload := req.Payload

	switch req.Cmd {
	case addr:
		h.HandleAddr(payload)
	case block:
		h.HandleBlock(payload)
	case inv:
		h.HandleInv(payload)
	case getBlocks:
		h.HandleGetBlocks(payload)
	case getData:
		h.HandleGetData(payload)
	case tx:
		h.HandleTx(payload)
	case ver:
		h.HandleVersion(payload)
	default:
		log.Printf("Unknown command %s", req.Cmd)
	}

	return nil
}

// handleSend handles sending data from one to other using default server.
func (h *HTTP) handleSend(ctx echo.Context) error {
	sendDTO := new(types.SendTokens)
	if err := ctx.Bind(sendDTO); err != nil {
		return errors.Wrap(err, "HandleSend: Failed to bind payload")
	}

	if !wallet.ValidateAddress(sendDTO.To) {
		log.Panic("Address to is not valid")
	}

	if !wallet.ValidateAddress(sendDTO.From) {
		log.Panic("Address from is not valid")
	}

	UTXOSet := blockchain.UTXOSet{Blockchain: h.chain}

	wallets, err := wallet.CreateWallets()
	blockchain.Handle(err)
	wallet.DeleteWalletLock()

	wlt := wallets.GetWallet(sendDTO.From)
	txn, err := blockchain.NewTransaction(wlt, sendDTO.To, sendDTO.Amount, &UTXOSet)

	if err != nil {
		return errors.Wrap(err, "HandleSend: Failed to create new transaction")
	}

	if sendDTO.MineNow {
		cbTx := blockchain.CoinbaseTx(sendDTO.From, "")
		txs := []*blockchain.Transaction{cbTx, txn}
		minedBlock := h.chain.MineBlock(txs)
		UTXOSet.Update(minedBlock)
	} else {
		SendTx(KnownNodes[0], txn)
		log.Println("send tx")
	}

	for _, node := range KnownNodes {
		if node != h.nodeID {
			SendVersion(node, h.chain)
		}
	}

	log.Println("Success!")

	return errors.Wrap(ctx.JSON(http.StatusOK, map[string]string{
		"message": "Success!",
	}), "HandleSend: Failed to send response")
}

func (h *HTTP) createWallet(ctx echo.Context) error {
	wallets, _ := wallet.CreateWallets()
	wlt := wallets.AddWallet()
	wallets.SaveFile()
	log.Printf("Your new address: %s\n", wlt.Address())

	return errors.Wrap(
		ctx.JSON(http.StatusCreated, map[string]interface{}{"address": string(wlt.Address())}),
		"failed to create wallet",
	)
}

func (h *HTTP) HandleAddr(payload interface{}) {
	addr := new(Addr)

	data, err := json.Marshal(payload)
	if err != nil {
		log.Panic(err)
	}

	if err := json.Unmarshal(data, addr); err != nil {
		log.Panic(err)
	}

	KnownNodes = append(KnownNodes, addr.AddrList...)
	log.Printf("There are %d known nodes now!\n", len(KnownNodes))
	RequestBlocks()
}

func RequestBlocks() {
	for _, node := range KnownNodes {
		SendGetBlocks(node)
	}
}

func (h *HTTP) HandleBlock(payload interface{}) {
	blockData := new(Block)

	data, err := json.Marshal(payload)
	if err != nil {
		log.Panic(err)
	}

	err = json.Unmarshal(data, blockData)
	if err != nil {
		log.Panic(err)
	}

	block := blockchain.Deserialize(blockData.Block)

	log.Println("Received a new block!")

	h.chain.AddBlock(block)

	log.Printf("Added block %x\n", block.Hash)
	log.Printf("Blocks in transit: %d\n", len(blocksInTransit))

	if len(blocksInTransit) > 0 {
		blockHash := blocksInTransit[0]
		blocksInTransit = blocksInTransit[1:]

		SendGetData(blockData.AddrFrom, "block", blockHash)
	} else {
		UTXOSet := blockchain.UTXOSet{Blockchain: h.chain}
		UTXOSet.Reindex()
	}
}

func (h *HTTP) HandleGetBlocks(payload interface{}) {
	blocks := new(GetBlocks)

	data, err := json.Marshal(payload)
	if err != nil {
		log.Panic(err)
	}

	err = json.Unmarshal(data, blocks)
	if err != nil {
		log.Panic(err)
	}

	blockHashes := h.chain.GetBlockHashes()
	SendInv(blocks.AddrFrom, "block", blockHashes)
}

func (h *HTTP) HandleGetData(payload interface{}) {
	getDataPayload := new(GetData)

	data, err := json.Marshal(payload)
	if err != nil {
		log.Panic(err)
	}

	err = json.Unmarshal(data, getDataPayload)
	if err != nil {
		log.Panic(err)
	}

	if getDataPayload.Type == "block" {
		block, err := h.chain.GetBlock(getDataPayload.ID)
		if err != nil {
			return
		}

		SendBlock(getDataPayload.AddrFrom, &block)
	}

	if getDataPayload.Type == "tx" {
		txID := hex.EncodeToString(getDataPayload.ID)
		tx := memoryPool[txID]

		SendTx(getDataPayload.AddrFrom, tx)
	}
}

func (h *HTTP) HandleVersion(payload interface{}) {
	ver := new(Version)

	data, err := json.Marshal(payload)
	if err != nil {
		log.Panic(err)
	}

	err = json.Unmarshal(data, ver)

	if err != nil {
		log.Panic(err)
	}

	myBestHeight := h.chain.GetBestHeight()
	foreignerBestHeight := ver.BestHeight

	if myBestHeight < foreignerBestHeight {
		SendGetBlocks(ver.AddrFrom)
	} else if myBestHeight > foreignerBestHeight {
		SendVersion(ver.AddrFrom, h.chain)
	}

	if !NodeIsKnown(ver.AddrFrom) {
		KnownNodes = append(KnownNodes, ver.AddrFrom)
	}
}

func NodeIsKnown(addr string) bool {
	for _, node := range KnownNodes {
		if node == addr {
			return true
		}
	}

	return false
}

func (h *HTTP) HandleTx(payload interface{}) {
	txn := new(Tx)

	data, err := json.Marshal(payload)
	if err != nil {
		log.Panic(err)
	}

	err = json.Unmarshal(data, txn)
	if err != nil {
		log.Panic(err)
	}

	txData := txn.Transaction
	transaction := blockchain.DeserializeTransaction(txData)
	memoryPool[hex.EncodeToString(transaction.ID)] = transaction

	log.Printf("Added transaction %x to mempool.\n", transaction.ID)

	if nodeAddress == KnownNodes[0] {
		for _, node := range KnownNodes {
			if node != nodeAddress && node != txn.AddrFrom {
				SendInv(node, "tx", [][]byte{transaction.ID})
			}
		}
	} else if len(memoryPool) >= 2 && len(KnownNodes) >= 2 {
		MineTx(h.chain)
	}
}

func MineTx(chain *blockchain.Blockchain) {
	var txs []*blockchain.Transaction

	for id := range memoryPool {
		log.Printf("Mining tx %x\n", memoryPool[id].ID)

		tx := memoryPool[id]
		if chain.VerifyTransaction(tx) {
			txs = append(txs, tx)
		}
	}

	if len(txs) == 0 {
		log.Println("All transactions are invalid! Waiting for new ones...")

		return
	}

	cbTx := blockchain.CoinbaseTx(minerAddress, "")
	txs = append(txs, cbTx)

	newBlock := chain.MineBlock(txs)
	UTXOSet := blockchain.UTXOSet{Blockchain: chain}
	UTXOSet.Reindex()

	log.Println("New block is mined!")

	for _, tx := range txs {
		txID := hex.EncodeToString(tx.ID)
		delete(memoryPool, txID)
	}

	for _, node := range KnownNodes {
		if node != nodeAddress {
			SendInv(node, "block", [][]byte{newBlock.Hash})
		}
	}

	if len(memoryPool) > 0 {
		MineTx(chain)
	}
}

func (h *HTTP) HandleInv(payload interface{}) {
	inventory := new(Inv)

	data, err := json.Marshal(payload)
	if err != nil {
		log.Panic(err)
	}

	err = json.Unmarshal(data, inventory)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Received inventory with %d %s\n", len(inventory.Items), inventory.Type)
	log.Printf("blocks in transit %d\n", len(blocksInTransit))

	if inventory.Type == "block" {
		blocksInTransit = inventory.Items

		blockHash := inventory.Items[0]
		SendGetData(inventory.AddrFrom, "block", blockHash)

		var newInTransit [][]byte

		for _, b := range blocksInTransit {
			if !bytes.Equal(b, blockHash) {
				newInTransit = append(newInTransit, b)
			}
		}

		blocksInTransit = newInTransit
	}

	if inventory.Type == "tx" {
		txID := inventory.Items[0]

		if memoryPool[hex.EncodeToString(txID)] == nil {
			SendGetData(inventory.AddrFrom, "tx", txID)
		}
	}
}

func (h *HTTP) getWallets(ctx echo.Context) error {
	wallets, _ := wallet.CreateWallets()
	addresses := wallets.GetAllAddresses()

	return errors.Wrap(
		ctx.JSON(http.StatusOK, map[string]interface{}{"addresses": addresses}),
		"error in getWallets")
}

func (h *HTTP) getChain(ctx echo.Context) error {
	chain := h.chain

	iterator := chain.Iterator()
	blocks := make([]*types.Block, 0)

	for {
		block := iterator.Next()
		log.Printf("Prev. hash: %x\n", block.PrevHash)
		log.Printf("Hash: %x\n", block.Hash)
		pow := blockchain.NewProof(block)
		isProofValid := pow.Validate()

		log.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))

		for _, tx := range block.Transactions {
			log.Println(tx)
		}

		log.Println()

		blocks = append(blocks, &types.Block{
			Timestamp: block.Timestamp,
			PrevHash:  fmt.Sprintf("%x", block.PrevHash),
			Hash:      fmt.Sprintf("%x", block.Hash),
			PoW:       isProofValid,
		})

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return errors.Wrap(ctx.JSON(http.StatusOK, types.Blockchain{Blocks: blocks}), "error in getBlockchain")
}

func (h *HTTP) getBalance(ctx echo.Context) error {
	address := ctx.Param("address")

	if !wallet.ValidateAddress(address) {
		log.Panic("Address is not valid")
	}

	chain := h.chain

	UTXOSet := blockchain.UTXOSet{Blockchain: chain}
	balance := 0

	pubKeyHash := wallet.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := UTXOSet.FindUTXO(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	return errors.Wrap(ctx.JSON(http.StatusOK, map[string]interface{}{"balance": balance}), "error in getBalance")
}
