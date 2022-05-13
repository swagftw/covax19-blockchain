package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
)

// ProofOfWork represents a proof-of-work algorithm
// which is nothing but algorithm to find a hash which satisfies a
// condition is we create a difficulty level which is the number of leading zeros required at in the hash
// for that we find a hash with bruteforce using a nonce which is a random number including other data in block.

const DIFFICULTY = 16

type ProofOfWork struct {
	Block  *Block
	Target *big.Int
}

// NewProof creates a new proof of work
func NewProof(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-DIFFICULTY))
	pow := &ProofOfWork{b, target}
	return pow
}

// InitData create the data from proof of work
func (pow *ProofOfWork) InitData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.Block.PrevHash,
			pow.Block.HashTransactions(),
			ToHex(int64(nonce)),
			ToHex(int64(DIFFICULTY)),
		},
		[]byte{},
	)
	return data
}

// Run runs the proof of work
func (pow *ProofOfWork) Run() (int, []byte) {
	var intHash big.Int
	var hash [32]byte

	nonce := 0

	for nonce < math.MaxInt64 {
		data := pow.InitData(nonce)
		hash = sha256.Sum256(data)
		fmt.Printf("\r%x", hash)
		intHash.SetBytes(hash[:])
		if intHash.Cmp(pow.Target) == -1 {
			break
		} else {
			nonce++
		}
	}

	fmt.Println()

	return nonce, hash[:]
}

// Validate validates the proof of work
func (pow *ProofOfWork) Validate() bool {
	var intHash big.Int
	data := pow.InitData(pow.Block.Nonce)
	hash := sha256.Sum256(data)
	intHash.SetBytes(hash[:])
	return intHash.Cmp(pow.Target) == -1
}

// ToHex converts int to hex
func ToHex(num int64) []byte {
	buffer := new(bytes.Buffer)
	err := binary.Write(buffer, binary.BigEndian, num)
	Handle(err)
	return buffer.Bytes()
}
