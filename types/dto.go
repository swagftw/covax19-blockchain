package types

type SendTokens struct {
	From             string `json:"from"`
	To               string `json:"to"`
	Amount           int    `json:"amount"`
	SkipBalanceCheck bool   `json:"skipBalanceCheck"`
}

type Block struct {
	PrevHash  string `json:"prevHash"`
	Hash      string `json:"hash"`
	PoW       bool   `json:"pow"`
	Timestamp int64  `json:"timestamp"`
}

type Blockchain struct {
	Blocks []*Block `json:"blocks"`
}

type CreateBlockchain struct {
	Address string `json:"address"`
}
