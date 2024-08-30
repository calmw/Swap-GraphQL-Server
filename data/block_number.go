package data

import (
	"context"
	"github.com/ethereum/go-ethereum/ethclient"
)

//type BlockNumberData struct {
//	Data struct {
//		LatestSaveBlockNumbers []BlockNumber `json:"latestSaveBlockNumbers"`
//	} `json:"data"`
//}
//
//type BlockNumber struct {
//	ID          string `json:"id"`
//	BlockNumber string `json:"blockNumber"`
//}
//
//func GetBlockNumberFromGraph() {
//
//	url := "http://127.0.0.1:8000/subgraphs/name/swap"
//	method := "POST"
//	for {
//		time.Sleep(time.Second * 5)
//		query := fmt.Sprintf(`{"query":"{ latestSaveBlockNumbers (first:5 ){ id blockNumber } }" }`)
//		payload := strings.NewReader(query)
//		client := &http.Client{Timeout: time.Second * 30}
//		req, err := http.NewRequest(method, url, payload)
//		if err != nil {
//			log.Println(err.Error())
//			continue
//		}
//		req.Header.Add("Content-Type", "application/json")
//		var res *http.Response
//		for k := 0; k < 30; k++ {
//			res, err = client.Do(req)
//			if err == nil {
//				break
//			}
//		}
//		if err != nil {
//			log.Println(err.Error())
//			continue
//		}
//		defer res.Body.Close()
//		body, err := io.ReadAll(res.Body)
//		if err != nil {
//			log.Println(err.Error())
//			continue
//		}
//		var record BlockNumberData
//		err = json.Unmarshal(body, &record)
//		if err != nil {
//			log.Println(err.Error())
//			continue
//		}
//		if len(record.Data.LatestSaveBlockNumbers) <= 0 {
//			log.Println("没有数据了")
//			continue
//		}
//
//		dataNum := len(record.Data.LatestSaveBlockNumbers)
//		for j := 0; j < dataNum; j++ {
//			r := record.Data.LatestSaveBlockNumbers[j]
//			InsertBlockNumber(r)
//		}
//	}
//}
//
//func InsertBlockNumber(event BlockNumber) {
//	key := []byte("block_number_event")
//	err := db.Pebble.Set(key, []byte(event.BlockNumber), pebble.ParseOrderData)
//	if err != nil {
//		log.Println(err)
//	}
//}

func Client() *ethclient.Client {
	client, err := ethclient.Dial("https://testnet-rpc.matchain.io")
	if err != nil {
		panic(err)
	}
	return client
}

func GetBlockNumber() (uint64, error) {
	cli := Client()
	return cli.BlockNumber(context.Background())
}
