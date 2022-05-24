package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/swagftw/covax19-blockchain/wallet"
	"log"
	"math/big"
	"strings"
	"time"
)

// Transaction represents a transaction
type Transaction struct {
	ID       []byte
	Inputs   []TxInput
	Outputs  []TxOutput
	LockTime int64
}

func (tx *Transaction) Serialize() []byte {
	var encoded bytes.Buffer
	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	return encoded.Bytes()
}

func (tx *Transaction) Hash() []byte {
	var hash [32]byte
	txCopy := *tx
	txCopy.ID = []byte{}
	hash = sha256.Sum256(txCopy.Serialize())
	return hash[:]
}

func NewTransaction(w *wallet.Wallet, to string, amount int, UTXO *UTXOSet) (*Transaction, error) {
	var inputs []TxInput
	var outputs []TxOutput

	pubKeyHash := wallet.PublicKeyToHash(w.PublicKey)
	acc, validOutputs := UTXO.FindSpendableOutputs(pubKeyHash, amount)

	if acc < amount {
		log.Println("Not enough funds")
		return nil, errors.New("not enough funds")
	}

	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		Handle(err)

		for _, out := range outs {
			input := TxInput{txID, out, nil, w.PublicKey}
			inputs = append(inputs, input)
		}
	}

	from := string(w.Address())

	outputs = append(outputs, *NewTXOutput(amount, to))

	if acc > amount {
		outputs = append(outputs, *NewTXOutput(acc-amount, from))
	}

	tx := Transaction{Inputs: inputs, Outputs: outputs, LockTime: time.Now().Unix()}
	tx.ID = tx.Hash()
	UTXO.Blockchain.SignTransaction(&tx, w.PrivateKey)

	return &tx, nil
}

// CoinbaseTx creates a new coinbase transaction
func CoinbaseTx(to, data string) *Transaction {
	if data == "" {
		randData := make([]byte, 20)
		_, err := rand.Read(randData)
		Handle(err)
		data = fmt.Sprintf("%x", randData)
	}
	txIn := TxInput{
		ID:        []byte{},
		Out:       -1,
		Signature: []byte(data),
	}

	txOut := NewTXOutput(20, to)

	tx := &Transaction{
		ID:       nil,
		Inputs:   []TxInput{txIn},
		Outputs:  []TxOutput{*txOut},
		LockTime: time.Now().Unix(),
	}

	tx.ID = tx.Hash()
	return tx
}

// IsCoinbase checks if a transaction is a coinbase
func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].ID) == 0 && tx.Inputs[0].Out == -1
}

func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if tx.IsCoinbase() {
		return
	}

	for _, in := range tx.Inputs {
		if prevTXs[hex.EncodeToString(in.ID)].ID == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()

	for inID, in := range txCopy.Inputs {
		prevTx := prevTXs[hex.EncodeToString(in.ID)]
		txCopy.Inputs[inID].Signature = nil
		txCopy.Inputs[inID].PublicKey = prevTx.Outputs[in.Out].PubKeyHash

		txCopy.ID = txCopy.Hash()
		txCopy.Inputs[inID].PublicKey = nil

		r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.ID)
		Handle(err)
		signature := append(r.Bytes(), s.Bytes()...)
		tx.Inputs[inID].Signature = signature
	}
}

func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	for _, in := range tx.Inputs {
		if prevTXs[hex.EncodeToString(in.ID)].ID == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	for inID, in := range txCopy.Inputs {
		prevTx := prevTXs[hex.EncodeToString(in.ID)]
		txCopy.Inputs[inID].Signature = nil
		txCopy.Inputs[inID].PublicKey = prevTx.Outputs[in.Out].PubKeyHash

		txCopy.ID = txCopy.Hash()
		txCopy.Inputs[inID].PublicKey = nil

		r := big.Int{}
		s := big.Int{}
		sigLen := len(in.Signature)

		r.SetBytes(in.Signature[:(sigLen / 2)])
		s.SetBytes(in.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(in.PublicKey)
		x.SetBytes(in.PublicKey[:(keyLen / 2)])
		y.SetBytes(in.PublicKey[(keyLen / 2):])

		rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}
		if !ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s) {
			return false
		}
	}

	return true
}

func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	for _, in := range tx.Inputs {
		inputs = append(inputs, TxInput{in.ID, in.Out, nil, nil})
	}
	for _, out := range tx.Outputs {
		outputs = append(outputs, TxOutput{out.Value, out.PubKeyHash})
	}

	txCopy := Transaction{
		ID:       tx.ID,
		Inputs:   inputs,
		Outputs:  outputs,
		LockTime: tx.LockTime,
	}
	return txCopy
}

func (tx Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.ID))
	for i, input := range tx.Inputs {
		lines = append(lines, fmt.Sprintf("\tInput %d:", i))
		lines = append(lines, fmt.Sprintf("\t\tTXID:     %x", input.ID))
		lines = append(lines, fmt.Sprintf("\t\tOut:       %d", input.Out))
		lines = append(lines, fmt.Sprintf("\t\tSignature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("\t\tPubKey:    %x", input.PublicKey))
	}

	for i, output := range tx.Outputs {
		lines = append(lines, fmt.Sprintf("\tOutput %d:", i))
		lines = append(lines, fmt.Sprintf("	\tValue:  %d", output.Value))
		lines = append(lines, fmt.Sprintf("	\tScript: %x", output.PubKeyHash))
	}

	return strings.Join(lines, "\n")
}
