package postgres

import (
	"context"

	"gorm.io/gorm"

	"github.com/swagftw/covax19-blockchain/utl/storage"
	tx "github.com/swagftw/covax19-blockchain/utl/transaction"
)

func NewPostgresTx(db *gorm.DB) tx.Transaction {
	return &transaction{db: db}
}

type transaction struct {
	db *gorm.DB
}

// Run is implementation of Transaction interface for gorm(postgres) transaction.
func (t transaction) Run(ctx context.Context, f func(context.Context) error) error {
	db := t.db.WithContext(ctx)

	if ctx.Value(storage.GormTxKey) != nil {
		if ctxDB, ok := ctx.Value(storage.GormTxKey).(*gorm.DB); ok {
			db = ctxDB
		}
	}

	return db.Transaction(func(tx *gorm.DB) error {
		ctx = context.WithValue(ctx, storage.GormTxKey, tx)

		return f(ctx)
	})
}
