package wallet

import (
	"log"
	"os"

	"github.com/mr-tron/base58"
)

func Base58Encode(input []byte) []byte {
	encode := base58.Encode(input)

	return []byte(encode)
}

func Base58Decode(input []byte) []byte {
	decode, err := base58.Decode(string(input))
	if err != nil {
		log.Panic()
	}

	return decode
}

func DeleteWalletLock() {
	if err := os.Remove(walletLock); err != nil {
		log.Panic(err)
	}
}
