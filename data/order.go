package data

import (
	"Swap-Server/db"
	"Swap-Server/models"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/cockroachdb/pebble"
	"gorm.io/gorm"
	"log"
	"time"
)

const StartBlockNumber = 1

type Order struct {
	FirstSwap bool
}

func NewOrder() *Order {
	return &Order{FirstSwap: true}
}

func (o *Order) Task() {
	o.Sync()

	ticker := time.NewTicker(time.Second * 10)
	for {
		select {
		case <-ticker.C:
			o.Sync()
		}
	}
}
func (o *Order) Sync() {
	swaps := o.getSwapTrace()
	order := models.Order{}
	for _, s := range swaps {
		//if s.TxHash == "0xe737c9e817b4ab477a40685671b7e6506a8c9a30f419404c9427e62546b4f6ea" {
		//if s.TxHash == "0xe737c9e817b4ab477a40685671b7e6506a8c9a30f419404c9427e62546b4f6ea" {
		//	fmt.Println("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
		//	fmt.Println(s.TxHash, s.Id, s.Amount, s.AmountIn)
		//}
		if s.From == "0x0000000000000000000000000000000000000000" || s.Src == "0x0000000000000000000000000000000000000000" {
			continue
		}
		ok, _ := o.getCurrentBlockNumberAndLogIndex(s.BlockNumber, s.LogIndex)
		if !ok {
			/// 转入的是token包括WBNB
			if o.FirstSwap {
				if s.SwapType == 3 && !o.isTransferToPair(s.Src) { // WBNB,来自用户（非pair）的转账
					order.User = s.Src
					order.FromAmount = s.Wad
				} else if s.SwapType == 4 && !o.isTransferToPair(s.From) { // Token,来自用户（非pair）的转账
					order.User = s.From
					order.FromAmount = s.Value
				} else {
					continue
				}
				order.FromToken = s.TokenAddress
				order.FromTokenSymbol = s.TokenSymbol
				order.UtcDateTime = s.UtcDateTime
				order.CreateTime = s.CreateTime
				order.TxHash = s.TxHash
				order.LogIndex = s.LogIndex
				order.BlockNumber = s.BlockNumber

				o.FirstSwap = false
			}
			if s.SwapType == 4 && !o.isTransferToPair(s.To) { // 没有转到pair合约,最后一步
				// 开始，转入到了pair合约
				order.ToAmount = s.Value
				order.ToTokenSymbol = s.TokenSymbol
				order.ToToken = s.TokenAddress
				o.Save(order)
				order = models.Order{}
				_ = o.setCurrentBlockNumberAndLogIndex(s.BlockNumber, s.LogIndex)
				o.FirstSwap = true
			} else if s.SwapType == 3 && !o.isTransferToPair(s.Dst) { // 没有转到pair合约,最后一步
				// 开始，转入到了pair合约
				order.ToAmount = s.Wad
				order.ToTokenSymbol = s.TokenSymbol
				order.ToToken = s.TokenAddress
				o.Save(order)
				order = models.Order{}
				_ = o.setCurrentBlockNumberAndLogIndex(s.BlockNumber, s.LogIndex)
				o.FirstSwap = true
			}
		}
	}
}

func (o *Order) Save(order models.Order) {
	key := []byte(fmt.Sprintf("order_create_%d_%d", order.BlockNumber, order.LogIndex))
	_, closer, err := db.Pebble.Get(key)
	if err == nil {
		closer.Close()
	}
	if errors.Is(err, pebble.ErrNotFound) {
		whereCondition := fmt.Sprintf("block_number=%d and log_index=%d", order.BlockNumber, order.LogIndex)
		err = db.PG.Model(models.Order{}).Where(whereCondition).First(&models.Order{}).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			db.PG.Model(models.Order{}).Create(&order)
			val := key
			err = db.Pebble.Set(key, val, pebble.Sync)
			if err != nil {
				log.Println(err)
				return
			}
			_ = o.setBlockNumber(uint64(order.BlockNumber))
		}
	}

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

func (o *Order) getCurrentBlockNumberAndLogIndex(blockNumber, logIndex int) (bool, error) {
	key := []byte(fmt.Sprintf("order_block_number_hash_%d_%d", blockNumber, logIndex))
	_, closer, err := db.Pebble.Get(key)
	if err == nil {
		_ = closer.Close()
		return true, nil
	}
	return false, err
}

func (o *Order) setCurrentBlockNumberAndLogIndex(blockNumber, logIndex int) error {
	key := []byte(fmt.Sprintf("order_block_number_hash_%d_%d", blockNumber, logIndex))
	val := []byte("ok")
	err := db.Pebble.Set(key, val, pebble.Sync)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
