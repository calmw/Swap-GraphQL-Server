package data

import (
	"Swap-Server/db"
	"Swap-Server/models"
	"encoding/binary"
	"fmt"
	"github.com/cockroachdb/pebble"
	"log"
)

const StartBlockNumber = 1

type Order struct {
}

func NewOrder() *Order {
	return &Order{}
}

func (o *Order) Create() {
	swaps := o.getSwapTrace()
	order := models.Order{}
	for _, s := range swaps {
		//if s.TxHash == "0xe737c9e817b4ab477a40685671b7e6506a8c9a30f419404c9427e62546b4f6ea" {
		//
		//	fmt.Println("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
		//	fmt.Println(s.TxHash, s.Id, s.Amount, s.AmountIn)
		//}
		/// 先按照 从存入token开始
		if s.SwapType == 4 {
			// transfers mint
			if s.From == "0x0000000000000000000000000000000000000000" {
				continue
			}
			// 没有转到pair合约
			if !o.isTransferToPair(s.From) {
				// 开始，转入到了pair合约
				order.User = s.From
				order.FromToken = s.TokenAddress
				order.FromAmount = s.Value
				order.UtcDateTime = s.UtcDateTime
				order.CreateTime = s.CreateTime
				order.TxHash = s.TxHash
				order.LogIndex = s.LogIndex
				order.BlockNumber = s.BlockNumber
			}
			// 没有转到pair合约
			if !o.isTransferToPair(s.To) {
				// 开始，转入到了pair合约
				order.ToAmount = s.Value
				order.ToToken = s.TokenAddress
				o.Save(order)
				order = models.Order{}
			}

		}
	}
}

func (o *Order) Save(order models.Order) {
	db.PG.Model(models.Order{}).Create(&order)
}

func (o *Order) isTransferToPair(to string) bool {
	var res bool
	for _, pair := range Pairs {
		if to == pair.Pair {
			res = true
		}
	}
	return res
}

func (o *Order) getSwapTrace() []models.SwapTrace {
	blockNumber := o.getBlockNumber()
	var swapTraces []models.SwapTrace

	whereCondition := fmt.Sprintf("block_number>%d", blockNumber)
	err := db.PG.Model(models.SwapTrace{}).Where(whereCondition).Order("block_number asc, log_index asc, tx_hash asc").Find(&swapTraces).Error
	if err != nil {
		log.Println(err)
	}

	return swapTraces
}

func (o *Order) getBlockNumber() uint64 {
	key := []byte("order_block_number")
	val, closer, err := db.Pebble.Get(key)
	if err == nil {
		_ = closer.Close()
		return binary.LittleEndian.Uint64(val)
	}
	return StartBlockNumber
}

func (o *Order) setBlockNumber(number uint64) error {
	key := []byte("order_block_number")
	bytesBuffer := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytesBuffer, number)
	err := db.Pebble.Set(key, bytesBuffer, pebble.Sync)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
