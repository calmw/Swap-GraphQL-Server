package models

import (
	"time"
)

type Order struct {
	Id              uint64    `gorm:"column:id;primaryKey" json:"id"`
	User            string    `gorm:"column:user" json:"user"`
	FromToken       string    `gorm:"column:from_token" json:"from_token"`
	ToToken         string    `gorm:"column:to_token" json:"to_token"`
	FromTokenSymbol string    `gorm:"column:from_token_symbol" json:"from_token_symbol"`
	ToTokenSymbol   string    `gorm:"column:to_token_symbol" json:"to_token_symbol"`
	Amount          string    `gorm:"column:amount" json:"amount"`
	CreateTime      time.Time `gorm:"column:create_time" json:"create_time"`
}

func NewOrder() *Order {
	return &Order{}
}

func (o *Order) AddFriend(uid, fid string) error {

	return nil
}
