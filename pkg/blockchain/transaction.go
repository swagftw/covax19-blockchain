package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/swagftw/covax19-blockchain/pkg/wallet"
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

	if err := enc.Encode(tx); err != nil {
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

func NewTransaction(w *wallet.Wallet, to string, amount int, UTXO *UTXOSet, skipBalanceCheck bool) (*Transaction, error) {
	inputs := make([]TxInput, 0)
	outputs := make([]TxOutput, 0)

	pubKeyHash := wallet.PublicKeyToHash(w.PublicKey)
	acc, validOutputs := UTXO.FindSpendableOutputs(pubKeyHash, amount)

	if !skipBalanceCheck {
		if acc < amount {
			log.Panic("Error: not enough funds")
		}
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

	newAmount := acc
	if !skipBalanceCheck {
		newAmount = acc - amount
	}

	outputs = append(outputs, *NewTXOutput(newAmount, from))

	tx := Transaction{
		ID:       nil,
		Inputs:   inputs,
		Outputs:  outputs,
		LockTime: time.Now().Unix(),
	}
	tx.ID = tx.Hash()
	UTXO.Blockchain.SignTransaction(&tx, w.PrivateKey)

	return &tx, nil
}

// CoinbaseTx creates a new coinbase transaction
func CoinbaseTx(to, data string) *Transaction {
	if data == "" {
		randData := make([]byte, 24)
		_, err := rand.Read(randData)
		Handle(err)

		data = fmt.Sprintf("%x", randData)
	}

	txin := TxInput{[]byte{}, -1, nil, []byte(data)}
	txout := NewTXOutput(20, to)

	tx := Transaction{Inputs: []TxInput{txin}, Outputs: []TxOutput{*txout}, LockTime: time.Now().Unix()}
	tx.ID = tx.Hash()

	return &tx
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

	for inId, in := range txCopy.Inputs {
		prevTX := prevTXs[hex.EncodeToString(in.ID)]
		txCopy.Inputs[inId].Signature = nil
		txCopy.Inputs[inId].PublicKey = prevTX.Outputs[in.Out].PubKeyHash

		dataToSign := fmt.Sprintf("%x\n", txCopy)

		r, s, err := ecdsa.Sign(rand.Reader, &privKey, []byte(dataToSign))
		Handle(err)

		signature := append(r.Bytes(), s.Bytes()...)

		tx.Inputs[inId].Signature = signature
		txCopy.Inputs[inId].PublicKey = nil
	}
}

func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	for _, in := range tx.Inputs {
		if prevTXs[hex.EncodeToString(in.ID)].ID == nil {
			log.Panic("Previous transaction not correct")
		}
	}

	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	for inID, in := range tx.Inputs {
		prevTx := prevTXs[hex.EncodeToString(in.ID)]
		txCopy.Inputs[inID].Signature = nil
		txCopy.Inputs[inID].PublicKey = prevTx.Outputs[in.Out].PubKeyHash

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

		dataToVerify := fmt.Sprintf("%x\n", txCopy)

		rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}

		if !ecdsa.Verify(&rawPubKey, []byte(dataToVerify), &r, &s) {
			return false
		}

		txCopy.Inputs[inID].PublicKey = nil
	}

	return true
}

func (tx *Transaction) TrimmedCopy() Transaction {
	inputs := make([]TxInput, 0)
	outputs := make([]TxOutput, 0)

	for _, in := range tx.Inputs {
		inputs = append(inputs, TxInput{in.ID, in.Out, nil, nil})
	}

	for _, out := range tx.Outputs {
		outputs = append(outputs, TxOutput{out.Value, out.PubKeyHash})
	}

	txCopy := Transaction{
		ID:      nil,
		Inputs:  inputs,
		Outputs: outputs,
	}

	return txCopy
}

func (tx Transaction) String() string {
	lines := make([]string, 0)

	lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.ID))
	for i, input := range tx.Inputs {
		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:     %x", input.ID))
		lines = append(lines, fmt.Sprintf("       Out:       %d", input.Out))
		lines = append(lines, fmt.Sprintf("       Signature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("       PubKey:    %x", input.PublicKey))
	}

	for i, output := range tx.Outputs {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %d", output.Value))
		lines = append(lines, fmt.Sprintf("       Script: %x", output.PubKeyHash))
	}

	return strings.Join(lines, "\n")
}