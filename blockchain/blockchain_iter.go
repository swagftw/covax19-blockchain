package blockchain

import (
	"github.com/dgraph-io/badger"
)

type Iterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

// Iterator returns an iterator over the blockchain
func (bc *Blockchain) Iterator() *Iterator {
	return &Iterator{bc.LastHash, bc.Database}
}

// Next returns the next block in the blockchain
func (iter *Iterator) Next() *Block {
	var block *Block

	err := iter.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(iter.CurrentHash)
		Handle(err)
		_ = item.Value(func(val []byte) error {
			block = Deserialize(val)

			return nil
		})

		return nil
	})
	Handle(err)

	iter.CurrentHash = block.PrevHash

	return block
}
