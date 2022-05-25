package types

type Wallet struct {
	Address string `json:"address,omitempty"`
	ID      string `json:"id,omitempty"`
}

type SendTokens struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Amount  int    `json:"amount"`
	MineNow bool   `json:"mineNow,omitempty"`
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