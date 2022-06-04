package transaction

import (
	"context"

	"gorm.io/gorm"

	"github.com/swagftw/covax19-blockchain/types"
)

type (
	// Repository represents the repository.
	Repository interface {
		SaveTransaction(ctx context.Context, transaction *types.Transaction) (*types.Transaction, error)
		GetTransactionFromAddress(ctx context.Context, address string) ([]*types.Transaction, error)
	}

	Transaction struct {
		ID          uint `gorm:"primaryKey"`
		FromAddress string
		ToAddress   string
		Amount      int
		gorm.Model
	}
)

func (t *Transaction) TableName() string {
	return "transactions.transactions"
}
