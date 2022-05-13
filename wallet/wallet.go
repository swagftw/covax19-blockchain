package wallet

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"golang.org/x/crypto/ripemd160"
	"log"
)

const (
	checkSumLength = 4
	version        = byte(0x00)
)

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

func (w *Wallet) Address() []byte {
	pubHash := PublicKeyToHash(w.PublicKey)

	versionedHash := append([]byte{version}, pubHash...)
	checksum := CheckSum(versionedHash)

	fullHash := append(versionedHash, checksum...)
	address := Base58Encode(fullHash)

	return address
}

func NewKeyPair() (ecdsa.PrivateKey, []byte) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Panic(err)
	}
	publicKey := append(privateKey.PublicKey.X.Bytes(), privateKey.PublicKey.Y.Bytes()...)
	return *privateKey, publicKey
}

func MakeWallet() *Wallet {
	private, public := NewKeyPair()
	wallet := Wallet{private, public}
	return &wallet
}

func PublicKeyToHash(publicKey []byte) []byte {
	pubHash := sha256.Sum256(publicKey)
	hasher := ripemd160.New()
	_, err := hasher.Write(pubHash[:])
	if err != nil {
		log.Panic(err)
	}
	return hasher.Sum(nil)
}

func CheckSum(data []byte) []byte {
	firstHash := sha256.Sum256(data)
	secondHash := sha256.Sum256(firstHash[:])

	return secondHash[:checkSumLength]
}

func ValidateAddress(address string) bool {
	pubKeyHash := Base58Decode([]byte(address))
	actualChecksum := pubKeyHash[len(pubKeyHash)-checkSumLength:]
	version := pubKeyHash[0]
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-checkSumLength]
	targetChecksum := CheckSum(append([]byte{version}, pubKeyHash...))
	return bytes.Compare(actualChecksum, targetChecksum) == 0
}
