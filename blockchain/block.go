package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
)

// Block represents each 'item' in the blockchain
type Block struct {
	Transactions []*Transaction
	PrevHash     []byte
	Hash         []byte
	Nonce        int
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
func CreateBlock(transactions []*Transaction, prevHash []byte) *Block {
	block := &Block{
		Transactions: transactions,
		PrevHash:     prevHash,
		Hash:         []byte{},
		Nonce:        0,
	}
	pow := NewProof(block)
	nonce, hash := pow.Run()
	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

// Genesis creates the first block in the blockchain
// which is called as the 'genesis block'
func Genesis(coinbase *Transaction) *Block {
	block := &Block{
		Transactions: []*Transaction{coinbase},
		PrevHash:     []byte{},
		Hash:         []byte{},
		Nonce:        0,
	}
	pow := NewProof(block)
	nonce, hash := pow.Run()
	block.Hash = hash[:]
	block.Nonce = nonce
	return block
}

func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte
	for _, transaction := range b.Transactions {
		txHashes = append(txHashes, transaction.Hash())
	}

	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))
	return txHash[:]
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
