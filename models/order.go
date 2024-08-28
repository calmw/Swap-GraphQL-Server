package models

type Order struct {
	Id              uint64 `gorm:"column:id;primaryKey" json:"id"`
	User            string `gorm:"column:user" json:"user"`
	FromToken       string `gorm:"column:from_token" json:"from_token"`
	ToToken         string `gorm:"column:to_token" json:"to_token"`
	FromTokenSymbol string `gorm:"column:from_token_symbol" json:"from_token_symbol"`
	ToTokenSymbol   string `gorm:"column:to_token_symbol" json:"to_token_symbol"`
	FromAmount      string `gorm:"column:from_amount" json:"from_amount"`
	ToAmount        string `gorm:"column:to_amount" json:"to_amount"`
	TxHash          string `gorm:"column:tx_hash" json:"tx_hash"`
	BlockNumber     int    `gorm:"column:block_number" json:"block_number"`
	LogIndex        int    `gorm:"column:log_index" json:"log_index"`
	UtcDateTime     string `gorm:"column:create_time" json:"utc_date_time"`
	CreateTime      int    `gorm:"column:create_time" json:"create_time"`
}

func NewOrder() *Order {
	return &Order{}
}

func (o *Order) AddFriend(uid, fid string) error {

	return nil
}
