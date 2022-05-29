package main

import (
	"github.com/swagftw/covax19-blockchain/pkg/blockchain/network"
	"log"
	"os"
)

func main() {
	defer os.Exit(0)
	nodeID := os.Getenv("NODE_ID")
	miner := os.Getenv("MINER_ADDR")

	if len(nodeID) == 0 {
		log.Panic("NODE_ID env is not set")
	}

	network.StartServer(nodeID, miner)
}
