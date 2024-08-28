package models

import (
	"time"
)

type Pair struct {
	Id          uint64    `gorm:"column:id;primaryKey" json:"id"`
	Pair        string    `gorm:"column:token_symbol" json:"pair"`
	PairName    string    `gorm:"column:token_symbol" json:"pair_name"`
	Token0      string    `gorm:"column:token_address" json:"token_0"`
	Token1      string    `gorm:"column:token_address" json:"token_1"`
	TxHash      string    `gorm:"column:tx_hash" json:"tx_hash"`
	BlockNumber string    `gorm:"column:block_number" json:"block_number"`
	LogIndex    string    `gorm:"column:log_index" json:"log_index"`
	CreateTime  time.Time `gorm:"column:create_time" json:"create_time"`
}
