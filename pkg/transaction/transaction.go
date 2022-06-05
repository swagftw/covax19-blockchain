package transaction

import (
	"context"
	"fmt"
	"net/http"

	"github.com/thoas/go-funk"

	"github.com/swagftw/covax19-blockchain/pkg/blockchain/network"
	"github.com/swagftw/covax19-blockchain/types"
	"github.com/swagftw/covax19-blockchain/utl/server"
	tx "github.com/swagftw/covax19-blockchain/utl/transaction"
)

type service struct {
	repo       Repository
	tx         tx.Transaction
	usrService types.UserService
}

func (s service) Send(ctx context.Context, dto *types.SendTokens) error {
	err := s.tx.Run(ctx, func(ctx context.Context) error {
		// get sender by address
		userFrom, err := s.usrService.GetUserByEmail(ctx, dto.From)
		if err != nil {
			return err
		}

		userTo, err := s.usrService.GetUserByEmail(ctx, dto.To)
		if err != nil {
			return err
		}

		if userFrom.Type == types.UserTypeGovernment {
			dto.SkipBalanceCheck = true
		}

		txn := &types.Transaction{
			FromAddress: userFrom.WalletAddress,
			ToAddress:   userTo.WalletAddress,
			Amount:      dto.Amount,
		}

		txn, err = s.repo.SaveTransaction(ctx, txn)
		if err != nil {
			return err
		}

		endpoint := fmt.Sprintf("http://%s/v1/transactions/send", network.KnownNodes[0])

		dto.From = userFrom.WalletAddress
		dto.To = userTo.WalletAddress

		_, err = server.SendRequest(http.MethodPost, endpoint, dto)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func (s service) SaveTransaction(ctx context.Context, transaction *types.Transaction) (*types.Transaction, error) {
	tx, err := s.repo.SaveTransaction(ctx, transaction)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func (s service) GetTransaction(ctx context.Context, address string) ([]*types.Transaction, error) {
	txns, err := s.repo.GetTransactionFromAddress(ctx, address)
	if err != nil {
		return nil, err
	}

	userAddresses := make([]string, 0)
	for _, txn := range txns {
		userAddresses = append(userAddresses, txn.FromAddress)
		userAddresses = append(userAddresses, txn.ToAddress)
	}

	users, err := s.usrService.GetUsersByAddresses(ctx, funk.UniqString(userAddresses))
	if err != nil {
		return nil, err
	}

	usersMap := funk.ToMap(users, "WalletAddress").(map[string]*types.User)
	for _, txn := range txns {
		txn.FromUser = usersMap[txn.FromAddress]
		txn.ToUser = usersMap[txn.ToAddress]
	}

	return txns, nil
}

func NewService(repo Repository, tx tx.Transaction, userService types.UserService) types.Service {
	return &service{
		repo:       repo,
		tx:         tx,
		usrService: userService,
	}
}
