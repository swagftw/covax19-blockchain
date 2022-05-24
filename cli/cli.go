package cli

import (
	"flag"
	"log"
	"os"
	"runtime"
	"strconv"

	"github.com/dgraph-io/badger"

	"github.com/swagftw/covax19-blockchain/blockchain"
	"github.com/swagftw/covax19-blockchain/network"
	"github.com/swagftw/covax19-blockchain/wallet"
)

type CommandLine struct{}

// PrintUsage prints the usage
func (cli *CommandLine) printUsage() {
	log.Println("Usage:")
	log.Println(" getbalance -address ADDRESS - get the balance of ADDRESS")
	log.Println(" createblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
	log.Println(" printchain - Print all the blocks of the blockchain")
	log.Println(" send -from FROM -to TO -amount AMOUNT -mine - Send AMOUNT of coins from FROM address to TO")
	log.Println(" createwallet - Create a new wallet")
	log.Println(" listaddresses - List all addresses in wallet file")
	log.Println(" reindexutxo - Rebuilds the UTXO set")
	log.Println(" startnode -miner ADDRESS - Start a node with ID specified in NODE_ID env. var. -miner enables mining")

}

// validateArgs validates the number of arguments
func (cli *CommandLine) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		runtime.Goexit()
	}
}

// printChain prints the Blockchain
func (cli *CommandLine) printChain(nodeID string) {
	chain, _ := blockchain.ContinueBlockchain(nodeID, network.KnownNodes[0])
	defer func(Database *badger.DB) {
		err := Database.Close()
		if err != nil {
			log.Panic(err)
		}
	}(chain.Database)

	iterator := chain.Iterator()

	for {
		block := iterator.Next()
		log.Printf("Prev. hash: %x\n", block.PrevHash)
		log.Printf("Hash: %x\n", block.Hash)
		pow := blockchain.NewProof(block)
		log.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))

		for _, tx := range block.Transactions {
			log.Println(tx)
		}

		log.Println()

		if len(block.PrevHash) == 0 {
			break
		}
	}
}

func (cli *CommandLine) createBlockchain(nodeID, address string) {
	if !wallet.ValidateAddress(address) {
		log.Panic("Address is not valid")
	}
	chain, _ := blockchain.InitBlockchain(address, nodeID)

	defer func(Database *badger.DB) {
		err := Database.Close()
		if err != nil {
			log.Panic(err)
		}
	}(chain.Database)

	UTXOSet := blockchain.UTXOSet{Blockchain: chain}

	UTXOSet.Reindex()
	log.Println("Blockchain created")
}

func (cli *CommandLine) getBalance(address, nodeID string) {
	if !wallet.ValidateAddress(address) {
		log.Panic("Address is not valid")
	}

	chain, _ := blockchain.ContinueBlockchain(nodeID, network.KnownNodes[0])

	defer func(Database *badger.DB) {
		err := Database.Close()
		if err != nil {
			log.Panic(err)
		}
	}(chain.Database)

	UTXOSet := blockchain.UTXOSet{Blockchain: chain}
	balance := 0

	pubKeyHash := wallet.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := UTXOSet.FindUTXO(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	log.Printf("Balance of '%s': %d\n", address, balance)
}

func (cli *CommandLine) send(from, to string, amount int, nodeID string, mineNow bool) {
	if !wallet.ValidateAddress(to) {
		log.Panic("Address to is not valid")
	}

	if !wallet.ValidateAddress(from) {
		log.Panic("Address from is not valid")
	}

	chain, _ := blockchain.ContinueBlockchain(nodeID, network.KnownNodes[0])

	defer func(Database *badger.DB) {
		err := Database.Close()
		if err != nil {
			log.Panic(err)
		}
	}(chain.Database)

	UTXOSet := blockchain.UTXOSet{Blockchain: chain}

	wallets, err := wallet.CreateWallets()
	blockchain.Handle(err)
	wlt := wallets.GetWallet(from)

	tx, err := blockchain.NewTransaction(wlt, to, amount, &UTXOSet)
	blockchain.Handle(err)
	if mineNow {
		cbTx := blockchain.CoinbaseTx(from, "")
		txs := []*blockchain.Transaction{cbTx, tx}
		block := chain.MineBlock(txs)
		UTXOSet.Update(block)
	} else {
		network.SendTx(network.KnownNodes[0], tx)
		log.Println("send tx")
	}

	log.Println("Success!")
}

func (cli *CommandLine) createWallet() {
	wallets, _ := wallet.CreateWallets()
	wlt := wallets.AddWallet()
	wallets.SaveFile()
	log.Printf("Your new address: %s\n", wlt.Address())
}

func (cli *CommandLine) listAddresses(nodeID string) {
	wallets, err := wallet.CreateWallets()
	if err != nil {
		log.Panic(err)
	}

	addresses := wallets.GetAllAddresses()

	for _, address := range addresses {
		log.Println(address)
	}
}

func (cli *CommandLine) reindexUTXO(nodeID string) {
	chain, _ := blockchain.ContinueBlockchain(nodeID, network.KnownNodes[0])
	defer chain.Database.Close()

	UTXO := blockchain.UTXOSet{Blockchain: chain}
	UTXO.Reindex()

	count := UTXO.CountTransactions()
	log.Printf("Done! There are %d transactions in the UTXO set.\n", count)
}

// StartNode starts a node
func (cli *CommandLine) StartNode(nodeID, minerAddr string) {
	log.Printf("Starting node %s\n", nodeID)

	if len(minerAddr) > 0 {
		if wallet.ValidateAddress(minerAddr) {
			log.Println("Mining is on. Address to receive rewards: ", minerAddr)
		} else {
			log.Panic("Wrong miner address!")
		}
	}

	network.StartServer(nodeID, minerAddr)
}

// Run starts the CLI
func (cli *CommandLine) Run() {
	cli.validateArgs()

	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		log.Printf("NODE_ID env. var is not set!")
		runtime.Goexit()
	}

	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	listAddressesCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)
	reindexUTXOCmd := flag.NewFlagSet("reindexutxo", flag.ExitOnError)
	startNodeCmd := flag.NewFlagSet("startnode", flag.ExitOnError)

	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendMine := sendCmd.Bool("mine", false, "Mine immediately on the same node")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")
	startNodeMiner := startNodeCmd.String("miner", "", "Enable mining mode and send reward to ADDRESS")

	switch os.Args[1] {
	case "reindexutxo":
		err := reindexUTXOCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			blockchain.Handle(err)
		}
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			blockchain.Handle(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			blockchain.Handle(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			blockchain.Handle(err)
		}

	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			blockchain.Handle(err)
		}

	case "listaddresses":
		err := listAddressesCmd.Parse(os.Args[2:])
		if err != nil {
			blockchain.Handle(err)
		}

	case "startnode":
		err := startNodeCmd.Parse(os.Args[2:])
		if err != nil {
			blockchain.Handle(err)
		}
	default:
		cli.printUsage()
		runtime.Goexit()
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			runtime.Goexit()
		}
		cli.getBalance(*getBalanceAddress, nodeID)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			runtime.Goexit()
		}
		cli.createBlockchain(nodeID, *createBlockchainAddress)
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			runtime.Goexit()
		}
		cli.send(*sendFrom, *sendTo, *sendAmount, nodeID, *sendMine)
	}

	if startNodeCmd.Parsed() {
		nodeID := os.Getenv("NODE_ID")
		if nodeID == "" {
			startNodeCmd.Usage()
			runtime.Goexit()
		}
		cli.StartNode(nodeID, *startNodeMiner)
	}

	if printChainCmd.Parsed() {
		cli.printChain(nodeID)
	}

	if createWalletCmd.Parsed() {
		cli.createWallet()
	}

	if listAddressesCmd.Parsed() {
		cli.listAddresses(nodeID)
	}

	if reindexUTXOCmd.Parsed() {
		cli.reindexUTXO(nodeID)
	}
}
