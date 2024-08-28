package data

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func GetSwapFromGraph() {
	url := "https://subgraph.intoverse.co/subgraphs/name/city_node"
	method := "POST"
	index := 0
	for {
		index++
		query := fmt.Sprintf(`{"query":"{ userLocationRecordV2S ( first:1000 skip:%d orderBy:ctime orderDirection:asc ){ id user countyId cityId location ctime txHash } }" }`, (index-1)*1000)
		payload := strings.NewReader(query)
		client := &http.Client{Timeout: time.Second * 30}
		req, err := http.NewRequest(method, url, payload)
		if err != nil {
			log.Logger.Error(err.Error())
			return
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
			log.Logger.Error(err.Error())
			return
		}
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		if err != nil {
			log.Logger.Error(err.Error())
			return
		}
		var record UserLocationRecordV2
		err = json.Unmarshal(body, &record)
		if err != nil {
			log.Logger.Error(err.Error())
			return
		}
		if len(record.Data.UserLocationRecordV2S) <= 0 {
			break
		}

		for j := 0; j < len(record.Data.UserLocationRecordV2S); j++ {
			r := record.Data.UserLocationRecordV2S[j]
			ParseUserLocationRecordEvent(r)
		}
	}
}

func SyncSwapFromGraph() {
	url := "https://subgraph.intoverse.co/subgraphs/name/city_node"
	method := "POST"

	query := fmt.Sprintf(`{"query":"{ userLocationRecordV2S ( first:1000 skip:0 orderBy:ctime orderDirection:desc ){ id user countyId cityId location ctime txHash } }" }`)
	payload := strings.NewReader(query)
	client := &http.Client{Timeout: time.Second * 30}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		log.Logger.Error(err.Error())
		return
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
		log.Logger.Error(err.Error())
		return
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Logger.Error(err.Error())
		return
	}
	var record UserLocationRecordV2
	err = json.Unmarshal(body, &record)
	if err != nil {
		log.Logger.Error(err.Error())
		return
	}
	if len(record.Data.UserLocationRecordV2S) <= 0 {
		return
	}

	for j := 0; j < len(record.Data.UserLocationRecordV2S); j++ {
		r := record.Data.UserLocationRecordV2S[j]
		ParseUserLocationRecordEvent(r)
	}

}

func ParseSwapEvent(r LocationRecordV2) {
	locationEncrypt := r.Location
	locationCode := utils2.ThreeDesDecrypt(locationEncrypt)
	code := strings.Split(locationCode, ",")
	if len(code) == 2 {
		code = append(code, "0")
	}
	if len(code) < 3 {
		log.Logger.Sugar().Warn("用户位置加密信息解析错误", r.User, locationCode, locationEncrypt, code)
		return
	}
	//if code[0] != "0" {
	//	log.Logger.Sugar().Warnln("国外用户位置信息", userAddress.String(), locationCode, locationEncrypt, code)
	//	code[2] = "0"
	//}
	// 国外用户
	codeSecond := code[1]
	codeSecondSlice := strings.Split(codeSecond, "")
	if code[0] != "0" && codeSecondSlice[0] != "0" {
		log.Logger.Sugar().Warn("国外用户位置信息", r.User, locationCode, locationEncrypt, code)
		//fmt.Println(code[2], 87679, len(code[2]))
		if len(code[2]) == 0 {
			code[2] = "0"
		}
	}
	// 容错，国内城市code首位是0的情况
	if codeSecondSlice[0] == "0" {
		// 获取城市code,根据区县code
		var areaCode models2.AreaCode
		whereCondition := fmt.Sprintf("ad_code=%s", code[2])
		err := db.Mysql.Table("area_code").Where(whereCondition).First(&areaCode).Error
		if err != nil {
			log.Logger.Sugar().Error(err)
			return
		}
		code[1] = fmt.Sprintf("%d", areaCode.CityCode)
	}

	//var timestamp int64
	//header, err := Cli.HeaderByNumber(context.Background(), big.NewInt(int64(logE.BlockNumber)))
	//if err == nil {
	//	timestamp = int64(header.Time)
	//}
	err := InsertUserLocation(r.User, r.CityID, r.CountyID, code, locationEncrypt, locationCode, r.Ctime)
	if err != nil {
		log.Logger.Sugar().Error(err)
		return
	}
}

func InsertSwap(userAddress, cityId, countyId string, code []string, locationEncrypt, locationCode string, dateTime string) error {

	var userLocation models2.UserLocation
	whereCondition := fmt.Sprintf("user='%s'", strings.ToLower(userAddress))
	err := db.Mysql.Table("user_location").Where(whereCondition).First(&userLocation).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 查询明文地址
		uri := fmt.Sprintf("https://wallet-api-v2.intowallet.io/api/v1/city_node/geographic_info?city_code=%s&ad_code=%s", code[1], code[2])
		log.Logger.Sugar().Info(uri)
		err, location := utils2.HttpGet(uri)
		if err != nil {
			log.Logger.Sugar().Error(err)
			return err
		}
		if strings.Contains(string(location), "There might be too much traffic") {
			time.Sleep(time.Second * 5)
			err, location = utils2.HttpGet(uri)
			if err != nil {
				log.Logger.Sugar().Error(err)
				return err
			}
		}
		var locationInfo LocationInfo
		err = json.Unmarshal(location, &locationInfo)
		if err != nil {
			log.Logger.Sugar().Error(err)
			return err
		}
		locationStr := locationInfo.Data.CountryName + " " + locationInfo.Data.CityName + " " + locationInfo.Data.AreaName
		if locationStr == "" {
			_ = RestoreUserLocation(strings.ToLower(userAddress))
		} else {
			db.Mysql.Model(&models2.UserLocation{}).Create(&models2.UserLocation{
				User:            strings.ToLower(userAddress),
				CountyId:        strings.ToLower(countyId),
				CityId:          strings.ToLower(cityId),
				LocationEncrypt: locationEncrypt,
				Location:        locationStr,
				Country:         locationInfo.Data.CountryName,
				City:            locationInfo.Data.CityName,
				County:          locationInfo.Data.AreaName,
				AreaCode:        locationCode,
				Ctime:           dateTime,
			})
		}
	} else if err == nil {
		// 查询明文地址
		uri := fmt.Sprintf("https://wallet-api-v2.intowallet.io/api/v1/city_node/geographic_info?city_code=%s&ad_code=%s", code[1], code[2])
		log.Logger.Sugar().Info(uri)
		err, location := utils2.HttpGet(uri)
		if err != nil {
			log.Logger.Sugar().Error(err)
			return err
		}
		fmt.Println(string(location), 5556666)
		if strings.Contains(string(location), "There might be too much traffic") {
			time.Sleep(time.Second * 5)
			err, location = utils2.HttpGet(uri)
			if err != nil {
				log.Logger.Sugar().Error(err)
				return err
			}
		}
		var locationInfo LocationInfo
		err = json.Unmarshal(location, &locationInfo)
		if err != nil {
			log.Logger.Sugar().Error(err)
			return err
		}
		locationStr := locationInfo.Data.CountryName + " " + locationInfo.Data.CityName + " " + locationInfo.Data.AreaName
		if locationStr == "" {
			_ = RestoreUserLocation(strings.ToLower(userAddress))
		} else {
			db.Mysql.Model(&models2.UserLocation{}).Where(whereCondition).Updates(map[string]interface{}{
				"county_id":        strings.ToLower(countyId),
				"city_id":          strings.ToLower(cityId),
				"location_encrypt": locationEncrypt,
				"location":         locationStr,
				"country":          locationInfo.Data.CountryName,
				"city":             locationInfo.Data.CityName,
				"county":           locationInfo.Data.AreaName,
				"area_code":        locationCode,
			})
		}

	}
	return nil
}
