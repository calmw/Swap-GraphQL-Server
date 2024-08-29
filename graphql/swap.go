package graphql

import "encoding/json"

type SwapOrder struct {
	Id        string `json:"id"`
	UtcTime   int64  `json:"utc_time"`
	Timestamp int64  `json:"timestamp"`
	FromToken string `json:"from_token"`
	AmountIn  string `json:"amount_in"`
	ToToken   string `json:"to_token"`
	AmountOut string `json:"amount_out"`
}

func (s SwapOrder) ToJson() string {
	marshal, err := json.Marshal(s)
	if err == nil {
		return string(marshal)
	}
	return ""
}
