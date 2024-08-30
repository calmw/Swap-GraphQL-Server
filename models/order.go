package models

import (
	"Swap-Server/db"
	"fmt"
	"reflect"
)

type Order struct {
	Id              uint64 `gorm:"column:id;primaryKey" json:"id"`
	User            string `gorm:"column:user" json:"user"`
	FromToken       string `gorm:"column:from_token" json:"from_token"`
	ToToken         string `gorm:"column:to_token" json:"to_token"`
	FromTokenSymbol string `gorm:"column:from_token_symbol" json:"from_token_symbol"`
	ToTokenSymbol   string `gorm:"column:to_token_symbol" json:"to_token_symbol"`
	FromAmount      string `gorm:"column:from_amount" json:"from_amount"`
	ToAmount        string `gorm:"column:to_amount" json:"to_amount"`
	TxHash          string `gorm:"column:tx_hash" json:"tx_hash"`
	BlockNumber     int    `gorm:"column:block_number" json:"block_number"`
	LogIndex        int    `gorm:"column:log_index" json:"log_index"`
	UtcDateTime     string `gorm:"column:utc_date_time" json:"utc_date_time"`
	CreateTime      int    `gorm:"column:create_time" json:"create_time"`
}

func NewOrder() *Order {
	return &Order{}
}

func (o *Order) Query(user, fromTokenSymbol, toTokenSymbol string) ([]map[string]interface{}, error) {
	var records []Order
	var result = make([]map[string]interface{}, 0)

	model := db.PG.Model(o)
	if len(user) > 0 {
		where := fmt.Sprintf(`"user"='%s'`, user)
		model.Where(where)
		//model.Where("user = ?", user)
	}
	if len(fromTokenSymbol) > 0 {
		where := fmt.Sprintf("from_token_symbol='%s'", fromTokenSymbol)
		model.Where(where)
	}
	if len(toTokenSymbol) > 0 {
		where := fmt.Sprintf("to_token_symbol='%s'", toTokenSymbol)
		model.Where(where)
	}
	err := model.Find(&records).Error
	if err != nil {
		return result, err
	}

	for _, record := range records {
		m, _ := StructToMap(record)
		//m := StructToMap(record)
		result = append(result, m)
	}

	return result, nil
}

// StructToMap 将结构体转换为map，键为json tag中的小写
func StructToMap(s interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %s", v.Kind())
	}
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		fieldName := t.Field(i).Name
		jsonTag := t.Field(i).Tag.Get("json")
		if jsonTag != "" {
			fieldName = jsonTag
		}
		result[fieldName] = v.Field(i).Interface()
	}
	return result, nil
}

//func StructToMap(s interface{}) map[string]interface{} {
//	t := reflect.TypeOf(s)
//	v := reflect.ValueOf(s)
//
//	var data = make(map[string]interface{})
//	for i := 0; i < t.NumField(); i++ {
//		data[t.Field(i).Name] = v.Field(i).Interface()
//	}
//	return data
//}
