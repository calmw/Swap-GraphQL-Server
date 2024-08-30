package data

import (
	"Swap-Server/db"
	"Swap-Server/models"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cockroachdb/pebble"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type SwapData struct {
	Data struct {
		Swaps []Swap `json:"swaps"`
	} `json:"data"`
}

type Swap struct {
	ID          string `json:"id"`
	Sender      string `json:"sender"`
	Pair        string `json:"pair"`
	PairName    string `json:"pairName"`
	Amount0In   string `json:"amount0In"`
	Amount1In   string `json:"amount1In"`
	Amount0Out  string `json:"amount0Out"`
	Amount1Out  string `json:"amount1Out"`
	To          string `json:"to"`
	BlockNumber string `json:"blockNumber"`
	LogIndex    string `json:"logIndex"`
	UtcTime     string `json:"utcTime"`
	Timestamp   string `json:"timestamp"`
	TxHash      string `json:"txHash"`
}

func GetSwapFromGraph() {
	var index uint64 = 1
	key := []byte("swap_event_index")
	val, closer, err := db.Pebble.Get(key)
	if err == nil {
		_ = closer.Close()
		index = binary.LittleEndian.Uint64(val)
	}
	url := "http://127.0.0.1:8000/subgraphs/name/swap"
	method := "POST"
	for {
		time.Sleep(time.Second * 10)
		query := fmt.Sprintf(`{"query":"{ swaps(first:50 skip:%d orderBy:blockNumber orderDirection:asc ){ id sender pair pairName amount0In amount1In amount0Out amount1Out to blockNumber logIndex utcTime timestamp txHash } }" }`,
			(index-1)*50,
		)
		payload := strings.NewReader(query)
		client := &http.Client{Timeout: time.Second * 30}
		req, err := http.NewRequest(method, url, payload)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		req.Header.Add("Content-Type", "application/json")
		var res *http.Response
		for k := 0; k < 30; k++ {
			res, err = client.Do(req)
			if err == nil {
				break
			}
		}
		if err != nil {
			log.Println(err.Error())
			continue
		}
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		var record SwapData
		err = json.Unmarshal(body, &record)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		if len(record.Data.Swaps) <= 0 {
			log.Println("没有数据了")
			time.Sleep(time.Second * 30)
			continue
		}

		dataNum := len(record.Data.Swaps)
		for j := 0; j < dataNum; j++ {
			r := record.Data.Swaps[j]
			InsertSwap(r)
		}
		if dataNum >= 50 {
			index++
		}
	}
}

func InsertSwap(event Swap) {
	var swapTrace models.SwapTrace
	key := []byte(fmt.Sprintf("swap_event_%s", event.ID))
	_, closer, err := db.Pebble.Get(key)
	if err == nil {
		closer.Close()
		return
	}
	if errors.Is(err, pebble.ErrNotFound) {
		blockNumber, _ := strconv.Atoi(event.BlockNumber)
		logIndex, _ := strconv.Atoi(event.LogIndex)
		whereCondition := fmt.Sprintf("block_number=%d and log_index=%d", blockNumber, logIndex)
		err = db.PG.Model(swapTrace).Where(whereCondition).First(&swapTrace).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			var amountIn string = "0"
			var amountOut string = "0"
			if event.Amount0In != "0" {
				amountIn = event.Amount0In
			} else if event.Amount1In != "0" {
				amountIn = event.Amount1In
			}
			if event.Amount0Out != "0" {
				amountOut = event.Amount0In
			} else if event.Amount1Out != "0" {
				amountOut = event.Amount1In
			}
			timestamp, _ := strconv.Atoi(event.Timestamp)

			err = db.PG.Model(&swapTrace).Create(&models.SwapTrace{
				TokenSymbol:  "",
				TokenAddress: "",
				To:           event.To,
				Pair:         event.Pair,
				PairName:     event.PairName,
				AmountIn:     amountIn,
				AmountOut:    amountOut,
				Receiver:     "",
				Amount:       "",
				Src:          "",
				Dst:          "",
				Wad:          "",
				From:         "",
				Value:        "",
				SwapType:     1,
				TxHash:       event.TxHash,
				BlockNumber:  blockNumber,
				LogIndex:     logIndex,
				UtcDateTime:  event.UtcTime,
				CreateTime:   timestamp,
			}).Error
			if err == nil {
				err = db.Pebble.Set(key, []byte(event.ID), pebble.Sync)
				if err != nil {
					log.Println(err)
					return
				}
			}
		}
	}
}
