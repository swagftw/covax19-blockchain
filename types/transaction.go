package types

import (
	"context"
	"errors"
	"time"
)

var ErrSavingTransaction = errors.New("error saving transaction")
var ErrGettingTransactions = errors.New("error getting transactions")
var ErrNotEnoughFunds = errors.New("not enough funds")

type (
	// Service represents transaction service.
	Service interface {
		Send(ctx context.Context, dto *SendTokens) error
		SaveTransaction(ctx context.Context, transaction *Transaction) (*Transaction, error)
		GetTransaction(ctx context.Context, address string) ([]*Transaction, error)
	}

	// Transaction represents transaction.
	Transaction struct {
		ID          uint      `json:"id"`
		FromAddress string    `json:"fromAddress"`
		ToAddress   string    `json:"toAddress"`
		Amount      int       `json:"amount"`
		FromUser    *User     `json:"fromUser"`
		ToUser      *User     `json:"toUser"`
		CreatedAt   time.Time `json:"createdAt"`
	}
)
