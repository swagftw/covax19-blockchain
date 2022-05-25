package network

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4/middleware"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/labstack/echo/v4"
	"github.com/vrecan/death/v3"

	"github.com/swagftw/covax19-blockchain/blockchain"
)

const (
	protocol = "tcp"
	version  = 1
)

var (
	nodeAddress     string
	minerAddress    string
	blocksInTransit [][]byte
	KnownNodes      = []string{
		"localhost:8080", // main node for chain operations
	}
	memoryPool = make(map[string]*blockchain.Transaction)
)

type command string

const (
	block     command = "block"
	tx        command = "tx"
	addr      command = "addr"
	getBlocks command = "getBlocks"
	getData   command = "getData"
	ver       command = "version"
	inv       command = "inv"
)

type Addr struct {
	AddrList []string `json:"addrList"`
}

type Block struct {
	AddrFrom string `json:"addrFrom"`
	Block    []byte `json:"block"`
}

type GetBlocks struct {
	AddrFrom string `json:"addrFrom"`
}

type GetData struct {
	AddrFrom string `json:"addrFrom"`
	Type     string `json:"type"`
	ID       []byte `json:"id"`
}

type Inv struct {
	AddrFrom string
	Type     string
	Items    [][]byte
}

type Tx struct {
	AddrFrom    string
	Transaction []byte
}

type Version struct {
	Version    int
	BestHeight int
	AddrFrom   string
}

type CmdRequest struct {
	Cmd     command     `json:"cmd"`
	Payload interface{} `json:"payload"`
}

type HTTP struct {
	chain  *blockchain.Blockchain
	nodeID string
}

type Send struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Amount  int    `json:"amount"`
	MineNow bool   `json:"mineNow,omitempty"`
}

// func ExtractCommand(request []byte) (command string) {
//	command = string(request[:commandLength])
//	return
//}

func SendData(addr string, request CmdRequest) {
	conn, err := net.Dial(protocol, addr)
	if err != nil {
		log.Printf("%s is not available\n", addr)

		updatedNodes := make([]string, 0)

		for _, node := range KnownNodes {
			if node != addr {
				updatedNodes = append(updatedNodes, node)
			}
		}

		KnownNodes = updatedNodes

		return
	}

	defer func(conn net.Conn) {
		err = conn.Close()
		if err != nil {
			log.Println(err)
		}
	}(conn)

	endpoint := fmt.Sprintf("http://%s/v1/cmd", addr)

	SendRequest(endpoint, request)
}

// func SendAddress(address string) {
//	nodes := Addr{KnownNodes}
//	nodes.AddrList = append(nodes.AddrList, nodeAddress)
//	request := CmdRequest{
//		Cmd:     "addr",
//		Payload: nodes,
//	}
//	SendData(address, request)
// }

func SendBlock(addr string, b *blockchain.Block) {
	data := Block{nodeAddress, b.Serialize()}
	request := CmdRequest{
		Cmd:     block,
		Payload: data,
	}
	SendData(addr, request)
}

func SendInv(address, kind string, items [][]byte) {
	inventory := Inv{nodeAddress, kind, items}
	request := CmdRequest{
		Cmd:     inv,
		Payload: inventory,
	}
	SendData(address, request)
}

func SendTx(addr string, t *blockchain.Transaction) {
	data := Tx{nodeAddress, t.Serialize()}
	request := CmdRequest{
		Cmd:     tx,
		Payload: data,
	}
	SendData(addr, request)
}

func SendVersion(addr string, chain *blockchain.Blockchain) {
	bestHeight := chain.GetBestHeight()
	payload := Version{version, bestHeight, nodeAddress}

	request := CmdRequest{
		Cmd:     ver,
		Payload: payload,
	}
	SendData(addr, request)
}

func SendGetBlocks(address string) {
	payload := GetBlocks{nodeAddress}
	request := CmdRequest{
		Cmd:     getBlocks,
		Payload: payload,
	}
	SendData(address, request)
}

func SendGetData(address, kind string, id []byte) {
	payload := GetData{nodeAddress, kind, id}
	request := CmdRequest{
		Cmd:     getData,
		Payload: payload,
	}
	SendData(address, request)
}

func CloseDB(chain *blockchain.Blockchain) {
	d := death.NewDeath(syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	d.WaitForDeathWithFunc(func() {
		defer os.Exit(1)
		defer runtime.Goexit()
		defer func(Database *badger.DB) {
			err := Database.Close()
			if err != nil {
				log.Printf("error closing badger db : %v", err)
			}
		}(chain.Database)
	})
}

// SendRequest sends a request to the node with the given address. (this is for inter-node communication)
func SendRequest(addr string, request CmdRequest) {
	var client http.Client

	data, err := json.Marshal(request)

	if err != nil {
		log.Panic(err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, addr, bytes.NewBuffer(data))
	if err != nil {
		log.Panic(err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		log.Panic(err)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Panic(err)
		}
	}(resp.Body)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Response: %s\n", string(body))
}

func StartServer(nodeID, minerAddr string) {
	nodeAddress = fmt.Sprintf("localhost:%s", nodeID)
	ech := echo.New()
	ech.Use(middleware.Logger(), middleware.Recover())
	log.Printf("Starting node %s\n", nodeAddress)

	minerAddress = minerAddr
	mainNodeID := GetMainNodeID()
	chain, err := blockchain.ContinueBlockchain(nodeID, mainNodeID)

	if errors.Is(err, blockchain.ErrNoBlockchain) {
		//chain = blockchain.CreateMainBlockchain(nodeID)
		log.Panic(err)
	}

	go CloseDB(chain)

	handler := HTTP{chain: chain, nodeID: nodeID}
	v1Group := ech.Group("/v1")
	v1Group.POST("/cmd", handler.handleCmd)

	chainGroup := v1Group.Group("/chain")
	chainGroup.GET("", handler.getChain)
	chainGroup.POST("/wallets", handler.createWallet)
	chainGroup.GET("/wallets", handler.getWallets)
	chainGroup.GET("/wallets/balance/:address", handler.getBalance)

	txGroup := v1Group.Group("/transactions")
	txGroup.POST("/send", handler.handleSend)

	errChan := make(chan error)

	go func(ech *echo.Echo) {
		err := ech.Start(":" + nodeID)
		errChan <- err
	}(ech)

	// wait for server to start
	for {
		_, err = net.Dial("tcp", fmt.Sprintf("localhost:%s", nodeID))
		if err == nil {
			break
		}

		time.Sleep(time.Millisecond * 100)
	}

	if nodeAddress != KnownNodes[0] {
		SendVersion(KnownNodes[0], chain)
	}

	go func() {
		for {
			log.Printf("mempool size: %d\n", len(memoryPool))
			time.Sleep(time.Second * 5)
		}
	}()

	log.Printf("%v", <-errChan)
}

func GetMainNodeID() string {
	return strings.Split(KnownNodes[0], ":")[1]
}

// func CmdToBytes(cmd string) []byte {
//	var bytes [commandLength]byte
//
//	for i, c := range cmd {
//		bytes[i] = byte(c)
//	}
//
//	return bytes[:]
//}

// func BytesToCmd(bytes []byte) string {
//	var cmd []byte
//
//	for _, b := range bytes {
//		if b != 0x0 {
//			cmd = append(cmd, b)
//		}
//	}
//
//	return log.Sprintf("%s", cmd)
//}

// func GobEncode(data interface{}) []byte {
//	var buff bytes.Buffer
//
//	enc := gob.NewEncoder(&buff)
//	err := enc.Encode(data)
//	if err != nil {
//		log.Panic(err)
//	}
//
//	return buff.Bytes()
//}
