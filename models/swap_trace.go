package models

type SwapTrace struct {
	Id uint64 `gorm:"column:id;primaryKey" json:"id"`

	// withdraw wbnbTransfers transfers
	TokenSymbol string `gorm:"column:token_symbol" json:"token_symbol"`

	// wbnbTransfers transfers
	TokenAddress string `gorm:"column:token_address" json:"token_address"`

	// swap transfers
	To string `gorm:"column:to" json:"to"`

	// swap
	Pair      string `gorm:"column:pair" json:"pair"`
	PairName  string `gorm:"column:pair_name" json:"pair_name"`
	AmountIn  string `gorm:"column:amount_in" json:"amount_in"`
	AmountOut string `gorm:"column:amount_out" json:"amount_out"`

	// withdraw
	Receiver string `gorm:"column:receiver" json:"receiver"`
	Amount   string `gorm:"column:amount" json:"amount"`

	// wbnbTransfers
	Src string `gorm:"column:src" json:"src"`
	Dst string `gorm:"column:dst" json:"dst"`
	Wad string `gorm:"column:wad" json:"wad"`

	// transfers
	From  string `gorm:"column:from" json:"from"`
	Value string `gorm:"column:value" json:"value"`

	//
	SwapType    int    `gorm:"column:swap_type" json:"swap_type"` // 1 swap; 2 withdraw 3 wbnbTransfers 4 transfers
	TxHash      string `gorm:"column:tx_hash" json:"tx_hash"`
	BlockNumber int    `gorm:"column:block_number" json:"block_number"`
	LogIndex    int    `gorm:"column:log_index" json:"log_index"`
	UtcDateTime string `gorm:"column:utc_date_time" json:"utc_date_time"`
	CreateTime  int    `gorm:"column:create_time" json:"create_time"`
}

func NewSwap() *SwapTrace {
	return &SwapTrace{}
}
