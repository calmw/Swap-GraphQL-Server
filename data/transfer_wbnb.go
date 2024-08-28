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

type TransferWBNBData struct {
	Data struct {
		WbnbTransfers []TransferWBNB `json:"wbnbTransfers"`
	} `json:"data"`
}

type TransferWBNB struct {
	ID           string `json:"id"`
	TokenSymbol  string `json:"tokenSymbol"`
	TokenAddress string `json:"tokenAddress"`
	Src          string `json:"src"`
	Dst          string `json:"dst"`
	Wad          string `json:"wad"`
	BlockNumber  string `json:"blockNumber"`
	LogIndex     string `json:"logIndex"`
	UtcTime      string `json:"utcTime"`
	Timestamp    string `json:"timestamp"`
	TxHash       string `json:"txHash"`
}

func GetTransferWBNBFromGraph() {
	var index uint64 = 1
	key := []byte("transferWBNB_event_index")
	val, closer, err := db.Pebble.Get(key)
	if err == nil {
		_ = closer.Close()
		index = binary.LittleEndian.Uint64(val)
	}
	url := "http://127.0.0.1:8000/subgraphs/name/swap"
	method := "POST"
	for {
		time.Sleep(time.Second * 10)
		query := fmt.Sprintf(`{"query":"{ wbnbTransfers ( first:50 skip:%d orderBy:blockNumber orderDirection:asc ){ id tokenSymbol tokenAddress src dst wad blockNumber logIndex utcTime timestamp txHash } }" }`, (index-1)*50)
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
		var record TransferWBNBData
		err = json.Unmarshal(body, &record)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		if len(record.Data.WbnbTransfers) <= 0 {
			log.Println("没有数据了")
			time.Sleep(time.Second * 30)
			continue
		}

		dataNum := len(record.Data.WbnbTransfers)
		for j := 0; j < dataNum; j++ {
			r := record.Data.WbnbTransfers[j]
			InsertTransferWBNB(r)
		}
		if dataNum >= 50 {
			index++
		}
	}
}

func InsertTransferWBNB(event TransferWBNB) {
	var swapTrace models.SwapTrace
	key := []byte(fmt.Sprintf("transferWBNB_event_%s", event.ID))
	_, closer, err := db.Pebble.Get(key)
	if err == nil {
		closer.Close()
	}
	if errors.Is(err, pebble.ErrNotFound) {
		blockNumber, _ := strconv.Atoi(event.BlockNumber)
		logIndex, _ := strconv.Atoi(event.LogIndex)
		whereCondition := fmt.Sprintf("block_number=%d and log_index=%d", blockNumber, logIndex)
		err = db.PG.Model(swapTrace).Where(whereCondition).First(&swapTrace).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			timestamp, _ := strconv.Atoi(event.Timestamp)
			err = db.PG.Model(&swapTrace).Create(&models.SwapTrace{
				TokenSymbol:  event.TokenSymbol,
				TokenAddress: event.TokenAddress,
				Dst:          event.Dst,
				Src:          event.Src,
				Wad:          event.Wad,
				SwapType:     3,
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
