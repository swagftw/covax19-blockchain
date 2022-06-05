package cli

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"

	"gorm.io/gorm"

	auth2 "github.com/swagftw/covax19-blockchain/pkg/auth"
	blockchain2 "github.com/swagftw/covax19-blockchain/pkg/blockchain"
	"github.com/swagftw/covax19-blockchain/pkg/blockchain/network"
	"github.com/swagftw/covax19-blockchain/pkg/user"
	wallet2 "github.com/swagftw/covax19-blockchain/pkg/wallet"
	"github.com/swagftw/covax19-blockchain/types"
	"github.com/swagftw/covax19-blockchain/utl/jwt"
	"github.com/swagftw/covax19-blockchain/utl/storage"
)

type CommandLine struct{}

func (cli *CommandLine) printUsage() {
	fmt.Println("Usage:")
	fmt.Println(" getbalance -address ADDRESS - get the balance for an address")
	fmt.Println(" createblockchain -address ADDRESS creates a blockchain and sends genesis reward to address")
	fmt.Println(" printchain - Prints the blocks in the chain")
	fmt.Println(" send -from FROM -to TO -amount AMOUNT -mine - Send amount of coins. Then -mine flag is set, mine off of this node")
	fmt.Println(" createwallet - Creates a new Wallet")
	fmt.Println(" listaddresses - Lists the addresses in our wallet file")
	fmt.Println(" reindexutxo - Rebuilds the UTXO set")
	fmt.Println(" startnode -miner ADDRESS - Start a node with ID specified in NODE_ID env. var. -miner enables mining")
}

func (cli *CommandLine) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		runtime.Goexit()
	}
}

func (cli *CommandLine) StartNode(nodeID, minerAddress string) {
	fmt.Printf("Starting Node %s\n", nodeID)

	if len(minerAddress) > 0 {
		if wallet2.ValidateAddress(minerAddress) {
			fmt.Println("Mining is on. Address to receive rewards: ", minerAddress)
		} else {
			log.Panic("Wrong miner address!")
		}
	}
	network.StartServer(nodeID, minerAddress)
}

func (cli *CommandLine) ReindexUTXO() {
	chain := blockchain2.ContinueBlockChain()
	defer chain.Database.Close()
	UTXOSet := blockchain2.UTXOSet{Blockchain: chain}
	UTXOSet.Reindex()

	count := UTXOSet.CountTransactions()
	fmt.Printf("Done! There are %d transactions in the UTXO set.\n", count)
}

func (cli *CommandLine) ListAddresses() {
	wallets, _ := wallet2.CreateWallets()
	addresses := wallets.GetAllAddresses()

	for _, address := range addresses {
		fmt.Println(address)
	}
}

func (cli *CommandLine) CreateWallet() {
	wallets, _ := wallet2.CreateWallets()
	wlt := wallets.AddWallet()
	wallets.SaveFile()

	log.Println("Your new address: ", string(wlt.Address()))
}

func (cli *CommandLine) PrintChain() {
	chain := blockchain2.ContinueBlockChain()
	defer chain.Database.Close()
	iter := chain.Iterator()

	for {
		block := iter.Next()

		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Printf("Prev. hash: %x\n", block.PrevHash)
		pow := blockchain2.NewProof(block)
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

func (cli *CommandLine) CreateBlockChain(address string) {
	if !wallet2.ValidateAddress(address) {
		log.Panic("Address is not Valid")
	}
	chain := blockchain2.InitBlockChain(address)
	defer chain.Database.Close()

	UTXOSet := blockchain2.UTXOSet{Blockchain: chain}
	UTXOSet.Reindex()

	fmt.Println("Finished!")
}

func (cli *CommandLine) GetBalance(address string) {
	if !wallet2.ValidateAddress(address) {
		log.Panic("Address is not Valid")
	}
	chain := blockchain2.ContinueBlockChain()
	UTXOSet := blockchain2.UTXOSet{Blockchain: chain}
	defer chain.Database.Close()

	balance := 0
	pubKeyHash := wallet2.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := UTXOSet.FindUnspentTransactions(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of %s: %d\n", address, balance)
}

func (cli *CommandLine) Send(from, to string, amount int, mineNow bool) {
	if !wallet2.ValidateAddress(to) {
		log.Panic("Address is not Valid")
	}
	if !wallet2.ValidateAddress(from) {
		log.Panic("Address is not Valid")
	}
	chain := blockchain2.ContinueBlockChain()
	UTXOSet := blockchain2.UTXOSet{Blockchain: chain}
	defer chain.Database.Close()

	wallets, err := wallet2.CreateWallets()
	if err != nil {
		log.Panic(err)
	}
	wallet2.DeleteWalletLock()
	wallet := wallets.GetWallet(from)

	tx, _ := blockchain2.NewTransaction(wallet, to, amount, &UTXOSet, true)

	if mineNow {
		cbTx := blockchain2.CoinbaseTx(from, "")
		txs := []*blockchain2.Transaction{cbTx, tx}
		block := chain.MineBlock(txs)
		UTXOSet.Update(block)
	} else {
		fmt.Println(hex.EncodeToString(tx.ID))
		network.SendTx(network.KnownNodes[0], tx)
		fmt.Println("send tx")
	}

	fmt.Println("Success!")
}

func (cli *CommandLine) Run() {
	cli.validateArgs()

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
	createBlockchainPassword := createBlockchainCmd.String("password", "", "password for the government account")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")
	sendMine := sendCmd.Bool("mine", false, "Mine immediately on the same node")
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
			log.Panic(err)
		}
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "startnode":
		err := startNodeCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "listaddresses":
		err := listAddressesCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
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
		cli.GetBalance(*getBalanceAddress)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			runtime.Goexit()
		}

		if *createBlockchainPassword == "" {
			createBlockchainCmd.Usage()
			runtime.Goexit()
		}

		cli.createGovernmentAccount(*createBlockchainPassword, *createBlockchainAddress)

		cli.CreateBlockChain(*createBlockchainAddress)
	}

	if printChainCmd.Parsed() {
		cli.PrintChain()
	}

	if createWalletCmd.Parsed() {
		cli.CreateWallet()
	}
	if listAddressesCmd.Parsed() {
		cli.ListAddresses()
	}
	if reindexUTXOCmd.Parsed() {
		cli.ReindexUTXO()
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			runtime.Goexit()
		}

		cli.Send(*sendFrom, *sendTo, *sendAmount, *sendMine)
	}

	if startNodeCmd.Parsed() {
		nodeID := os.Getenv("NODE_ID")
		if nodeID == "" {
			startNodeCmd.Usage()
			runtime.Goexit()
		}
		cli.StartNode(nodeID, *startNodeMiner)
	}
}

func (cli *CommandLine) createGovernmentAccount(password, address string) {
	gdb, err := storage.NewPostgresDB()
	if err != nil {
		log.Panic(err)
	}

	err = gdb.Transaction(func(txn *gorm.DB) error {
		// create password.
		pwd := &user.Password{
			Password: password,
		}
		err = txn.Save(pwd).Error
		if err != nil {
			return err
		}

		usr := &user.User{
			Name:          "Central Government",
			Email:         "government@gov.in",
			Type:          types.UserTypeGovernment,
			WalletAddress: address,
			Verified:      true,
			PasswordID:    pwd.ID,
		}

		err = txn.Model(usr).Create(usr).Error
		if err != nil {
			return err
		}

		u := &types.User{
			ID:            usr.ID,
			Name:          usr.Name,
			Email:         usr.Email,
			Type:          usr.Type,
			WalletAddress: usr.WalletAddress,
			Verified:      usr.Verified,
		}

		// create user.
		jwtService, _ := jwt.New()
		accessToken, _ := jwtService.GenerateAccessToken(u)
		refreshToken := jwtService.GenerateRefreshToken(accessToken)

		auth := &auth2.Tokens{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			Identifier:   u.Email,
		}

		err = txn.Save(auth).Error
		if err != nil {
			return err
		}

		return err
	})
	if err != nil {
		log.Panic(err)
	}
}
