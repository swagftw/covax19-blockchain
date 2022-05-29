package wallet

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"io/ioutil"
	"log"
	"os"
	"time"
)

const walletFile = "./tmp/wallets.data"
const walletLock = "./tmp/wallets.LOCK"

type Wallets struct {
	Wallets map[string]*Wallet
}

func (ws *Wallets) SaveFile() {
	walletFileExists := false
	if _, err := os.Stat(walletFile); err == nil {
		walletFileExists = true
	}
	// make sure lock file exists
	if _, err := os.Stat(walletLock); os.IsNotExist(err) && walletFileExists {
		log.Panic(err)
	}

	var content bytes.Buffer

	gob.Register(elliptic.P256())
	encoder := gob.NewEncoder(&content)

	err := encoder.Encode(ws)
	if err != nil {
		log.Panic(err)
	}

	err = ioutil.WriteFile(walletFile, content.Bytes(), 0600) //nolint:gomnd
	if err != nil {
		log.Panic(err)
	}

	DeleteWalletLock()
}

func (ws *Wallets) LoadFile() error {
	// if LOCK file does not exist, create it
	// if LOCK exists, wait for it to be deleted
	for {
		if _, err := os.Stat(walletLock); err != nil {
			_, err = os.Create(walletLock)
			if err != nil {
				return err
			}

			break
		}

		log.Println("waiting for wallet lock file to be deleted")
		time.Sleep(time.Millisecond * 10) //nolint:gomnd
	}

	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return err
	}

	var wallets Wallets

	fileContent, err := ioutil.ReadFile(walletFile)
	if err != nil {
		return err
	}

	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&wallets)

	if err != nil {
		return err
	}

	ws.Wallets = wallets.Wallets

	return nil
}

func CreateWallets() (*Wallets, error) {
	var wallets Wallets

	wallets.Wallets = make(map[string]*Wallet)

	if err := wallets.LoadFile(); err != nil {
		return &wallets, err
	}

	return &wallets, nil
}

func (ws *Wallets) AddWallet() *Wallet {
	wallet := MakeWallet()
	address := string(wallet.Address())

	ws.Wallets[address] = wallet

	return wallet
}

func (ws *Wallets) GetWallet(address string) *Wallet {
	return ws.Wallets[address]
}

func (ws *Wallets) GetAllAddresses() []string {
	addresses := make([]string, 0)

	for address := range ws.Wallets {
		addresses = append(addresses, address)
	}

	DeleteWalletLock()

	return addresses
}
