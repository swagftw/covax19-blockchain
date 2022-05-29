package types

import "context"

// ChainService provides access to the blockchain.
type ChainService interface {
	GetBlocksInChain(ctx context.Context, start, end int) ([]*Block, error)
	SendTokens(ctx context.Context, from, to, amount string, mineNow bool) error
}
