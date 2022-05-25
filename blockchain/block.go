package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"time"
)

// Block represents each 'item' in the blockchain
type Block struct {
	Timestamp    int64
	Transactions []*Transaction
	Difficulty   int
	PrevHash     []byte
	Hash         []byte
	Nonce        int
	Height       int
}

// DeriveHash returns the hash of the block
func (b *Block) DeriveHash() {
	// join concatenates two given bytes of data with a separator
	info := bytes.Join([][]byte{b.HashTransactions(), b.Hash}, []byte{})
	// actually a more complicated method is used to calculate the hash in actual blockchain
	hash := sha256.Sum256(info)
	b.Hash = hash[:]
}

// CreateBlock creates a new block
func CreateBlock(transactions []*Transaction, prevHash []byte, height int) *Block {
	block := &Block{
		Transactions: transactions,
		PrevHash:     prevHash,
		Hash:         []byte{},
		Nonce:        0,
		Timestamp:    time.Now().Unix(),
		Height:       height,
		Difficulty:   DIFFICULTY,
	}
	pow := NewProof(block)
	nonce, hash := pow.Run()
	block.Hash = hash
	block.Nonce = nonce

	return block
}

// Genesis creates the first block in the blockchain
// which is called as the 'genesis block'
func Genesis(coinbase *Transaction) *Block {
	return CreateBlock([]*Transaction{coinbase}, []byte{}, 0)
}

func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte
	for _, transaction := range b.Transactions {
		txHashes = append(txHashes, transaction.Serialize())
	}
	tree := NewMerkleTree(txHashes)
	return tree.RootNode.Data
}

// Serialize serializes the block into a byte array
func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(b)
	Handle(err)
	return result.Bytes()
}

// Deserialize deserializes the block from a byte array
func Deserialize(data []byte) *Block {
	block := new(Block)
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(block)
	Handle(err)
	return block
}
