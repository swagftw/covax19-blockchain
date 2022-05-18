package cli

import (
	"flag"
	"fmt"
	"github.com/swagftw/covax19-blockchain/blockchain"
	"github.com/swagftw/covax19-blockchain/network"
	"github.com/swagftw/covax19-blockchain/wallet"
	"log"
	"os"
	"runtime"
	"strconv"
)

type CommandLine struct{}

// PrintUsage prints the usage
func (cli *CommandLine) printUsage() {
	fmt.Println("Usage:")
	fmt.Println(" getbalance -address ADDRESS - get the balance of ADDRESS")
	fmt.Println(" createblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println(" printchain - Print all the blocks of the blockchain")
	fmt.Println(" send -from FROM -to TO -amount AMOUNT -mine - Send AMOUNT of coins from FROM address to TO")
	fmt.Println(" createwallet - Create a new wallet")
	fmt.Println(" listaddresses - List all addresses in wallet file")
	fmt.Println(" reindexutxo - Rebuilds the UTXO set")
	fmt.Println("startnode -miner ADDRESS - Start a node with ID specified in NODE_ID env. var. -miner enables mining")

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
	chain := blockchain.ContinueBlockchain(nodeID)
	defer chain.Database.Close()
	iterator := chain.Iterator()
	for {
		block := iterator.Next()
		fmt.Printf("Prev. hash: %x\n", block.PrevHash)
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := blockchain.NewProof(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}
		fmt.Println()
		if len(block.PrevHash) == 0 {
			break
		}
	}
}

func (cli *CommandLine) createBlockchain(nodeID, address string) {
	if !wallet.ValidateAddress(address) {
		log.Panic("Address is not valid")
	}
	chain := blockchain.InitBlockchain(address, nodeID)
	defer chain.Database.Close()
	UTXOSet := blockchain.UTXOSet{Blockchain: chain}
	UTXOSet.Reindex()
	fmt.Println("Blockchain created")
}

func (cli *CommandLine) getBalance(address string) {
	if !wallet.ValidateAddress(address) {
		log.Panic("Address is not valid")
	}
	chain := blockchain.ContinueBlockchain(address)
	defer chain.Database.Close()
	UTXOSet := blockchain.UTXOSet{Blockchain: chain}

	balance := 0

	pubKeyHash := wallet.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := UTXOSet.FindUTXO(pubKeyHash)
	for _, out := range UTXOs {
		balance += out.Value
	}
	fmt.Printf("Balance of '%s': %d\n", address, balance)
}

func (cli *CommandLine) send(from, to string, amount int, nodeID string, mineNow bool) {
	if !wallet.ValidateAddress(to) {
		log.Panic("Address to is not valid")
	}
	if !wallet.ValidateAddress(from) {
		log.Panic("Address from is not valid")
	}
	chain := blockchain.ContinueBlockchain(nodeID)
	defer chain.Database.Close()
	UTXOSet := blockchain.UTXOSet{Blockchain: chain}

	wallets, err := wallet.CreateWallets(nodeID)
	blockchain.Handle(err)
	wallet := wallets.GetWallet(from)

	tx := blockchain.NewTransaction(wallet, to, amount, &UTXOSet)
	if mineNow {
		cbTx := blockchain.CoinbaseTx(from, "")
		txs := []*blockchain.Transaction{cbTx, tx}
		block := chain.MineBlock(txs)
		UTXOSet.Update(block)
	} else {
		network.SendTx(network.KnownNodes[0], tx)
		fmt.Println("send tx")
	}

	fmt.Println("Success!")
}

func (cli *CommandLine) createWallet(nodeID string) {
	wallets, _ := wallet.CreateWallets(nodeID)
	address := wallets.AddWallet()
	wallets.SaveFile(nodeID)
	fmt.Printf("Your new address: %s\n", address)
}

func (cli *CommandLine) listAddresses(nodeID string) {
	wallets, err := wallet.CreateWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	addresses := wallets.GetAllAddresses()
	for _, address := range addresses {
		fmt.Println(address)
	}
}

func (cli *CommandLine) reindexUTXO(nodeID string) {
	chain := blockchain.ContinueBlockchain(nodeID)
	defer chain.Database.Close()

	UTXO := blockchain.UTXOSet{Blockchain: chain}
	UTXO.Reindex()

	count := UTXO.CountTransactions()
	fmt.Printf("Done! There are %d transactions in the UTXO set.\n", count)
}

func (cli *CommandLine) StartNode(nodeID, minerAddr string) {
	fmt.Printf("Starting node %s\n", nodeID)

	if len(minerAddr) > 0 {
		if wallet.ValidateAddress(minerAddr) {
			fmt.Println("Mining is on. Address to receive rewards: ", minerAddr)
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
		fmt.Printf("NODE_ID env. var is not set!")
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
		cli.getBalance(*getBalanceAddress)
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

		cli.StartNode(nodeID, *createBlockchainAddress)
	}

	if printChainCmd.Parsed() {
		cli.printChain(nodeID)
	}

	if createWalletCmd.Parsed() {
		cli.createWallet(nodeID)
	}

	if listAddressesCmd.Parsed() {
		cli.listAddresses(nodeID)
	}

	if reindexUTXOCmd.Parsed() {
		cli.reindexUTXO(nodeID)
	}
}
