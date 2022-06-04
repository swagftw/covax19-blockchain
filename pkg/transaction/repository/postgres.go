package repository

import (
	"context"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"

	"github.com/swagftw/covax19-blockchain/pkg/transaction"
	"github.com/swagftw/covax19-blockchain/types"
	"github.com/swagftw/covax19-blockchain/utl/storage"
)

type repo struct {
	db *gorm.DB
}

func (r repo) SaveTransaction(ctx context.Context, tran *types.Transaction) (*types.Transaction, error) {
	db := storage.GetGormDBFromContext(ctx, r.db)
	txn := new(transaction.Transaction)

	err := db.Transaction(func(tx *gorm.DB) error {
		err := copier.Copy(txn, tran)
		if err != nil {
			return types.ErrCopy
		}
		err = tx.Save(txn).Error
		if err != nil {
			return types.ErrSavingTransaction
		}
		err = copier.Copy(tran, txn)
		if err != nil {
			return types.ErrCopy
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return tran, nil
}

func (r repo) GetTransactionFromAddress(ctx context.Context, address string) ([]*types.Transaction, error) {
	db := storage.GetGormDBFromContext(ctx, r.db)
	txns := make([]*transaction.Transaction, 0)

	resp := make([]*types.Transaction, 0)
	err := db.Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&transaction.Transaction{}).Where("to_address = ? OR from_address = ?", address, address).Find(&txns).Error
		if err != nil {
			return types.ErrGettingTransactions
		}

		err = copier.Copy(&resp, &txns)
		if err != nil {
			return types.ErrCopy
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return resp, nil
}

func NewRepo(db *gorm.DB) transaction.Repository {
	return &repo{
		db: db,
	}
}
