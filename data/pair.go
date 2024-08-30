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

type PairData struct {
	Data struct {
		Pairs []Pair `json:"pairCreateds"`
	} `json:"data"`
}

type Pair struct {
	ID          string `json:"id"`
	Pair        string `json:"pair"`
	Token0      string `json:"token0"`
	Token1      string `json:"token1"`
	PairName    string `json:"pairName"`
	BlockNumber string `json:"blockNumber"`
	LogIndex    string `json:"logIndex"`
	UtcTime     string `json:"utcTime"`
	Timestamp   string `json:"timestamp"`
	TxHash      string `json:"txHash"`
}

var Pairs []models.Pair

func GetPairFromGraph() {
	var index uint64 = 1
	key := []byte("pair_event_index")
	val, closer, err := db.Pebble.Get(key)
	if err == nil {
		_ = closer.Close()
		index = binary.LittleEndian.Uint64(val)
	}
	url := "http://127.0.0.1:8000/subgraphs/name/swap"
	method := "POST"
	for {
		time.Sleep(time.Second * 10)
		query := fmt.Sprintf(`{"query":"{ pairCreateds(first:50 skip:%d orderBy:blockNumber orderDirection:asc ) { id pair token0 token1 pair pairName blockNumber logIndex utcTime timestamp txHash } }" }`, (index-1)*50)
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
		var record PairData
		err = json.Unmarshal(body, &record)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		if len(record.Data.Pairs) <= 0 {
			log.Println("没有数据了")
			time.Sleep(time.Second * 30)
			continue
		}

		dataNum := len(record.Data.Pairs)
		for j := 0; j < dataNum; j++ {
			r := record.Data.Pairs[j]
			InsertPair(r)
		}
		if dataNum >= 50 {
			index++
		}
	}
}

func InsertPair(event Pair) {
	var pair models.Pair
	key := []byte(fmt.Sprintf("pair_event_%s", event.ID))
	_, closer, err := db.Pebble.Get(key)
	if err == nil {
		closer.Close()
		return
	}
	if errors.Is(err, pebble.ErrNotFound) {
		blockNumber, _ := strconv.Atoi(event.BlockNumber)
		logIndex, _ := strconv.Atoi(event.LogIndex)
		whereCondition := fmt.Sprintf("block_number=%d and log_index=%d", blockNumber, logIndex)
		err = db.PG.Model(pair).Where(whereCondition).First(&pair).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			timestamp, _ := strconv.Atoi(event.Timestamp)
			err = db.PG.Model(&pair).Create(&models.Pair{
				Pair:        event.Pair,
				PairName:    event.PairName,
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

func UpdatePair() {
	var pairs []models.Pair
	ticker := time.NewTicker(time.Minute * 30)
	for {
		select {
		case <-ticker.C:
			err := db.PG.Model(models.Pair{}).Find(&pairs).Error
			if err == nil {
				Pairs = pairs
			} else {
				log.Println(err)
			}
		}
	}
}

func SetPair() {
	var pairs []models.Pair
	err := db.PG.Model(models.Pair{}).Find(&pairs).Error
	if err == nil {
		Pairs = pairs
	} else {
		log.Println(err)
	}
}
