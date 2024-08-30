package data

import (
	"Swap-Server/db"
	"Swap-Server/models"
	"errors"
	"fmt"
	"github.com/cockroachdb/pebble"
	"gorm.io/gorm"
	"log"
	"time"
)

const MaxFailedTimes = 10

var RouterAddress = "0xc5d4d7b9a90c060f1c7d389bc3a20eeb382aa665"

type SwapChanData struct {
	Swaps  []models.SwapTrace
	TxHash string
}

var SwapChain = make(chan SwapChanData)

type Order struct{}

func NewOrder() *Order {
	return &Order{}
}

func (o *Order) Task() {
	go func() {
		ticker := time.NewTicker(time.Second * 3)
		for {
			select {
			case <-ticker.C:
				swaps, txHash := o.getSwapTrace()
				if len(swaps) > 0 {
					SwapChain <- SwapChanData{
						Swaps:  swaps,
						TxHash: txHash,
					}
				}
			}
		}
	}()

	for data := range SwapChain {
		go o.ParseOrderData(data.Swaps, data.TxHash)
	}
}

func (o *Order) ParseOrderData(swaps []models.SwapTrace, txHash string) {
	if len(swaps) == 0 {
		return
	}

	for i := 0; i < MaxFailedTimes; i++ {
		order := models.Order{}
		firstSwap := true
		ok := false
		for _, s := range swaps {
			if s.From == "0x0000000000000000000000000000000000000000" || s.Src == "0x0000000000000000000000000000000000000000" {
				return
			}
			//fmt.Println(o.isPair(s.To) || o.isPair(s.Dst), s.BlockNumber, s.LogIndex, 6666)
			/// 转入的是token包括WBNB
			if firstSwap {
				if s.SwapType == 3 { // WBNB,来自用户（非pair）的转账
					//if s.Src == RouterAddress {
					if !o.isPair(s.Src) {
						order.FromAmount = s.Wad
						if s.TokenSymbol == "WBNB" {
							order.FromTokenSymbol = "BNB"
							order.FromToken = ""
						}
					}
				} else if s.SwapType == 4 && !o.isPair(s.From) { // Token,来自用户（非pair）的转账
					order.User = s.From
					order.FromAmount = s.Value
					order.FromTokenSymbol = s.TokenSymbol
				} else {
					firstSwap = false
					continue
				}
				order.FromToken = s.TokenAddress
				order.UtcDateTime = s.UtcDateTime
				order.CreateTime = s.CreateTime
				order.TxHash = s.TxHash
				order.LogIndex = s.LogIndex
				order.BlockNumber = s.BlockNumber

				firstSwap = false
				continue
			}
			if s.SwapType == 4 && !o.isPair(s.To) { // 没有转到pair合约,最后一步
				// 开始，转入到了pair合约
				if len(order.User) <= 0 {
					order.User = s.To
				}
				order.ToTokenSymbol = s.TokenSymbol
				order.ToToken = s.TokenAddress
				order.ToAmount = s.Value
				o.Save(order)
				ok = true
				break
			} else if s.SwapType == 3 { // 没有转到pair合约,最后一步
				if s.Dst == RouterAddress {
					continue
				}
				if !o.isPair(s.Dst) {
					// 开始，转入到了pair合约
					order.ToAmount = s.Wad
					order.ToTokenSymbol = s.TokenSymbol
					order.ToToken = s.TokenAddress
					o.Save(order)
					ok = true
					break
				}
			} else if s.SwapType == 2 { // withdraw
				tokenSymbol := s.TokenSymbol
				tokenAddress := s.TokenAddress
				if s.TokenSymbol == "WBNB" {
					tokenSymbol = "BNB"
					tokenAddress = ""
				}
				order.ToAmount = s.Amount
				order.ToTokenSymbol = tokenSymbol
				order.ToToken = tokenAddress
				ok = true
				break
			}
		}
		time.Sleep(time.Second * 3)
		if !ok {
			o.addFailedTxHash(txHash)
		}
	}
}

// 每次一个交易
func (o *Order) getSwapTrace() ([]models.SwapTrace, string) {
	var swapTrace models.SwapTrace
	swapTraces := make([]models.SwapTrace, 0)
	txHash := o.getLatestTxHash()
	err := db.PG.Model(models.SwapTrace{}).Where("tx_hash=?", txHash).Order("block_number desc, log_index desc").First(&swapTrace).Error
	if err != nil {
		return swapTraces, txHash
	}

	whereCondition := fmt.Sprintf("(block_number=%d and log_index>%d) or (block_number>%d)", swapTrace.BlockNumber, swapTrace.LogIndex, swapTrace.BlockNumber)
	swapTrace = models.SwapTrace{}
	err = db.PG.Model(models.SwapTrace{}).Where(whereCondition).Order("block_number asc, log_index asc, tx_hash asc").First(&swapTrace).Error
	if err != nil {
		log.Println(err)
		return swapTraces, txHash
	}

	txHash = swapTrace.TxHash

	err = db.PG.Model(models.SwapTrace{}).Where("tx_hash=?", swapTrace.TxHash).Order("block_number asc, log_index asc, tx_hash asc").Find(&swapTraces).Error
	if err == nil {
		o.setLatestTxHash(txHash)
	} else {
		log.Println(err)
	}

	return swapTraces, txHash
}

func (o *Order) Save(order models.Order) {
	key := []byte(fmt.Sprintf("order_create_%d_%d", order.BlockNumber, order.LogIndex))
	_, closer, err := db.Pebble.Get(key)
	if err == nil {
		closer.Close()
		return
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
		}
	}
}

func (o *Order) isPair(to string) bool {
	var res bool
	for _, pair := range Pairs {
		if to == pair.Pair {
			res = true
			break
		}
	}
	return res
}

func (o *Order) getLatestTxHash() string {
	key := []byte("latest_txHash")
	val, closer, err := db.Pebble.Get(key)
	if err == nil {
		_ = closer.Close()
		return string(val)
	}
	var order models.Order
	err = db.PG.Model(models.Order{}).Order("block_number desc, log_index desc").First(&order).Error
	if err == nil {
		return order.TxHash
	}
	var swapTrace models.SwapTrace
	err = db.PG.Model(models.SwapTrace{}).Order("block_number asc, log_index asc").First(&swapTrace).Error
	if err == nil {
		return swapTrace.TxHash
	}
	return ""
}

func (o *Order) setLatestTxHash(txHash string) {
	key := []byte("latest_txHash")
	val := []byte(txHash)
	_ = db.Pebble.Set(key, val, pebble.Sync)
}

func (o *Order) addFailedTxHash(txHash string) {
	var failedTx models.FailedTx
	err := db.PG.Model(models.FailedTx{}).Where("tx_hash=?", txHash).First(&failedTx).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		db.PG.Model(models.FailedTx{}).Create(&models.FailedTx{TxHash: txHash, FailedTimes: 1})
	} else if err == nil {
		if failedTx.FailedTimes >= MaxFailedTimes {
			o.setLatestTxHash(txHash)
		} else {
			db.PG.Model(models.FailedTx{}).Where("tx_hash=?", txHash).Update("failed_times", failedTx.FailedTimes+1)
		}
	}
}
