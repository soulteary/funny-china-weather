package weather

import (
	"errors"
)

type responseIPIP struct {
	Ret  string `json:"ret"`
	Data struct {
		IP       string   `json:"ip"`
		Location []string `json:"location"`
	} `json:"data"`
}

func GetMyIPLocation() (string, error) {
	response := new(responseIPIP)
	err := getJSON("https://myip.ipip.net/json", 3, response)
	if err != nil {
		return "", err
	}

	locationData := response.Data.Location
	if locationData[0] != "中国" {
		return "", errors.New("非境内IP地址，放弃继续解析 :(")
	}

	province := locationData[1]
	city := locationData[2]
	if province == "北京" || province == "上海" || province == "重庆" || province == "天津" {
		return province + "市", nil
	}
	return (province + "省" + city + "市"), nil
}
