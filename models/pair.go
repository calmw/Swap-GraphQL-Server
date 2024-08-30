package models

type Pair struct {
	Id          uint64 `gorm:"column:id;primaryKey" json:"id"`
	Pair        string `gorm:"column:pair" json:"pair"`
	PairName    string `gorm:"column:pair_name" json:"pair_name"`
	TxHash      string `gorm:"column:tx_hash" json:"tx_hash"`
	BlockNumber int    `gorm:"column:block_number" json:"block_number"`
	LogIndex    int    `gorm:"column:log_index" json:"log_index"`
	UtcDateTime string `gorm:"column:create_time" json:"utc_date_time"`
	CreateTime  int    `gorm:"column:create_time" json:"create_time"`
}
