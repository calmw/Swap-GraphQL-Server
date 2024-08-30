package models

type FailedTx struct {
	Id          uint64 `gorm:"column:id;primaryKey" json:"id"`
	TxHash      string `gorm:"column:tx_hash" json:"tx_hash"`
	FailedTimes int    `gorm:"column:failed_times" json:"failed_times"`
	UtcDateTime string `gorm:"column:create_time" json:"utc_date_time"`
}
