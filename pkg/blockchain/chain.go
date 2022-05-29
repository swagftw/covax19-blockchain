package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/swagftw/covax19-blockchain/pkg/wallet"

	"github.com/dgraph-io/badger"
)

const (
	DBPath      = "./tmp/blocks_%s"
	GenesisData = "Data For Genesis Transaction data"
)

var (
	ErrBlockchainExists = errors.New("blockchain already exists")
	ErrNoBlockchain     = errors.New("blockchain not found")
)

// Blockchain is a series of validated Blocks and is the actual blockchain that is stored.
type Blockchain struct {
	LastHash []byte
	Database *badger.DB
}

// InitBlockchain creates the genesis block in the blockchain and creates the blockchain.
func InitBlockchain(address, nodeID string) (*Blockchain, error) {
	path := fmt.Sprintf(DBPath, nodeID)
	if DBExists(path) {
		log.Println("blockchain exists")

		return nil, ErrBlockchainExists
	}

	var lastHash []byte

	// Open the database
	opts := badger.DefaultOptions(path)
	bdDB, err := openDB(path, opts)
	Handle(err)

	// if the database is not present create genesis block and store it in the database
	// else continue with the blockchain
	err = bdDB.Update(func(txn *badger.Txn) error {
		cbtx := CoinbaseTx(address, GenesisData)
		genesis := Genesis(cbtx)
		log.Println("Creating and storing genesis block")
		err = txn.Set(genesis.Hash, genesis.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), genesis.Hash)

		lastHash = genesis.Hash

		return err
	})
	Handle(err)

	return &Blockchain{lastHash, bdDB}, nil
}

func ContinueBlockchain(nodeID, mainNodeID string) (*Blockchain, error) {
	path := fmt.Sprintf(DBPath, nodeID)
	if !DBExists(path) {
		err := copyFromMainBlockchain(nodeID, mainNodeID)
		if err != nil {
			return nil, err
		}
	}

	var lastHash []byte

	// Open the database
	opts := badger.DefaultOptions(path)
	db, err := openDB(path, opts)
	Handle(err)

	err = db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		err = item.Value(func(val []byte) error {
			lastHash = val

			return nil
		})

		return err
	})
	Handle(err)

	chain := &Blockchain{lastHash, db}

	return chain, nil
}

func CreateMainBlockchain(mainNodeID string) *Blockchain {
	wallets, _ := wallet.CreateWallets()
	wlt := wallets.AddWallet()

	wallets.SaveFile()

	address := string(wlt.Address())

	if !wallet.ValidateAddress(address) {
		log.Panic("Address is not valid")
	}

	chain, _ := InitBlockchain(address, mainNodeID)
	UTXOSet := UTXOSet{Blockchain: chain}

	UTXOSet.Reindex()

	return chain
}

// MineBlock adds a new block to the blockchain
func (bc *Blockchain) MineBlock(transactions []*Transaction) *Block {
	var lastHash []byte

	var lastHeight int

	for _, tx := range transactions {
		if !bc.VerifyTransaction(tx) {
			log.Panic("Invalid Transaction")
		}
	}

	err := bc.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		err = item.Value(func(val []byte) error {
			lastHash = val

			return nil
		})
		if err != nil {
			log.Panic(err)
		}

		item, err = txn.Get(lastHash)
		Handle(err)
		var lastBlockData []byte
		_ = item.Value(func(val []byte) error {
			lastBlockData = val

			return nil
		})

		lastBlock := Deserialize(lastBlockData)

		lastHeight = lastBlock.Height

		return err
	})
	Handle(err)

	newBlock := CreateBlock(transactions, lastHash, lastHeight+1)

	err = bc.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), newBlock.Hash)

		bc.LastHash = newBlock.Hash

		return err
	})
	Handle(err)

	return newBlock
}

func (bc *Blockchain) AddBlock(block *Block) {
	err := bc.Database.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get(block.Hash); err == nil {
			return nil
		}
		err := txn.Set(block.Hash, block.Serialize())
		Handle(err)

		item, err := txn.Get([]byte("lh"))
		Handle(err)
		var lastHash []byte
		err = item.Value(func(val []byte) error {
			lastHash = val

			return nil
		})
		Handle(err)
		item, err = txn.Get(lastHash)
		Handle(err)
		err = item.Value(func(val []byte) error {
			lastBlock := Deserialize(val)
			if block.Height > lastBlock.Height {
				err = txn.Set([]byte("lh"), block.Hash)
				Handle(err)
				bc.LastHash = block.Hash
			}

			return err
		})

		return nil
	})
	Handle(err)
}

func (bc *Blockchain) GetBlock(blockHash []byte) (Block, error) {
	var block Block

	err := bc.Database.View(func(txn *badger.Txn) error {
		if item, err := txn.Get(blockHash); err != nil {
			return errors.New("block not found")
		} else {
			err := item.Value(func(val []byte) error {
				block = *Deserialize(val)

				return nil
			})

			return err
		}
	})

	return block, err
}

func (bc *Blockchain) GetBestHeight() int {
	var lastBlock Block

	var lastHash []byte

	err := bc.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		err = item.Value(func(val []byte) error {
			lastHash = val

			return nil
		})
		Handle(err)
		item, err = txn.Get(lastHash)
		Handle(err)
		err = item.Value(func(val []byte) error {
			lastBlock = *Deserialize(val)

			return nil
		})

		return err
	})
	Handle(err)

	return lastBlock.Height
}

func (bc *Blockchain) GetBlockHashes() [][]byte {
	var blocks [][]byte

	iter := bc.Iterator()

	for {
		block := iter.Next()
		blocks = append(blocks, block.Hash)

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return blocks
}

func (bc *Blockchain) FindUTXOs() map[string]TxOutputs {
	UTXO := make(map[string]TxOutputs)
	spentTXOs := make(map[string][]int)

	iter := bc.Iterator()

	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Outputs {
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}
			if !tx.IsCoinbase() {
				for _, in := range tx.Inputs {
					inTxID := hex.EncodeToString(in.ID)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Out)
				}
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return UTXO
}

func (bc *Blockchain) FindTransaction(id []byte) (Transaction, error) {
	iter := bc.Iterator()

	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, id) == 0 {
				return *tx, nil
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("transaction is not found")
}

func (bc *Blockchain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	for _, in := range tx.Inputs {
		prevTX, err := bc.FindTransaction(in.ID)
		Handle(err)
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sign(privKey, prevTXs)
}

func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	prevTXs := make(map[string]Transaction)

	for _, in := range tx.Inputs {
		prevTX, err := bc.FindTransaction(in.ID)
		Handle(err)

		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs)
}

func DeserializeTransaction(data []byte) *Transaction {
	transaction := new(Transaction)
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(transaction)
	Handle(err)

	return transaction
}

func retry(dir string, originalOpts badger.Options) (*badger.DB, error) {
	lockPath := filepath.Join(dir, "LOCK")
	if err := os.Remove(lockPath); err != nil {
		return nil, fmt.Errorf(`removing "LOCK": %s`, err)
	}
	retryOpts := originalOpts
	retryOpts.Truncate = true
	return badger.Open(retryOpts)
}

func openDB(dir string, opts badger.Options) (*badger.DB, error) {
	if db, err := badger.Open(opts); err != nil {
		if strings.Contains(err.Error(), "LOCK") {
			fmt.Println("db locked, retrying")
			if db, err = retry(dir, opts); err == nil {
				log.Println("database unlocked, continuing")
				return db, nil
			}
			log.Println("could not unlock database:", err)
		}
		return nil, err
	} else {
		return db, nil
	}
}

func DBExists(path string) bool {
	if _, err := os.Stat(path + "/MANIFEST"); os.IsNotExist(err) {
		return false
	}
	return true
}

func copyFromMainBlockchain(node string, mainNodeID string) error {
	mainNodeDBPath := fmt.Sprintf(DBPath, mainNodeID)
	if !DBExists(mainNodeDBPath) {
		log.Println("blockchain not found")

		return ErrNoBlockchain
	}

	sourceDir, err := ioutil.ReadDir(mainNodeDBPath)
	Handle(err)

	destDirPath := fmt.Sprintf(DBPath, node)
	err = os.Mkdir(destDirPath, 0755)
	Handle(err)

	for _, fileInfo := range sourceDir {
		if fileInfo.Name() == "LOCK" {
			continue
		}
		sourceFilePath := filepath.Join(mainNodeDBPath, fileInfo.Name())
		destFilePath := filepath.Join(destDirPath, fileInfo.Name())
		copyFile(sourceFilePath, destFilePath)
	}
	return nil
}

func copyFile(sourceFilePath, destFilePath string) {
	file, err := os.Open(sourceFilePath)
	Handle(err)
	defer file.Close()
	fileData, err := io.ReadAll(file)
	Handle(err)
	err = os.WriteFile(destFilePath, fileData, 0644)
	Handle(err)
}
