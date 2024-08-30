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

type WithDrawData struct {
	Data struct {
		Withdraws []Withdraw `json:"withdraws"`
	} `json:"data"`
}

type Withdraw struct {
	ID          string `json:"id"`
	TokenSymbol string `json:"tokenSymbol"`
	Receiver    string `json:"receiver"`
	Amount      string `json:"amount"`
	BlockNumber string `json:"blockNumber"`
	LogIndex    string `json:"logIndex"`
	UtcTime     string `json:"utcTime"`
	Timestamp   string `json:"timestamp"`
	TxHash      string `json:"txHash"`
}

func GetWithDrawFromGraph() {
	var index uint64 = 1
	key := []byte("withdraw_event_index")
	val, closer, err := db.Pebble.Get(key)
	if err == nil {
		_ = closer.Close()
		index = binary.LittleEndian.Uint64(val)
	}
	url := "http://127.0.0.1:8000/subgraphs/name/swap"
	method := "POST"
	for {
		time.Sleep(time.Second * 10)
		query := fmt.Sprintf(`{"query":"{ withdraws(first:50 skip:%d orderBy:blockNumber orderDirection:asc ){ id tokenSymbol receiver amount blockNumber logIndex utcTime timestamp txHash } }" }`, (index-1)*50)
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
		var record WithDrawData
		err = json.Unmarshal(body, &record)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		if len(record.Data.Withdraws) <= 0 {
			log.Println("没有数据了")
			time.Sleep(time.Second * 30)
			continue
		}

		dataNum := len(record.Data.Withdraws)
		for j := 0; j < dataNum; j++ {
			r := record.Data.Withdraws[j]
			InsertWithdraw(r)
		}
		if dataNum >= 50 {
			index++
		}
	}
}

func InsertWithdraw(event Withdraw) {
	var swapTrace models.SwapTrace
	key := []byte(fmt.Sprintf("withdraw_event_%s", event.ID))
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
			timestamp, _ := strconv.Atoi(event.Timestamp)
			err = db.PG.Model(&swapTrace).Create(&models.SwapTrace{
				TokenSymbol: event.TokenSymbol,
				Receiver:    event.Receiver,
				Amount:      event.Amount,
				SwapType:    2,
				TxHash:      event.TxHash,
				BlockNumber: blockNumber,
				LogIndex:    logIndex,
				UtcDateTime: event.UtcTime,
				CreateTime:  timestamp,
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
