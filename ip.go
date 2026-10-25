package weather

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type responseIPIP struct {
	Ret  string `json:"ret"`
	Data struct {
		IP       string   `json:"ip"`
		Location []string `json:"location"`
	} `json:"data"`
}

// GetMyIPLocation returns the province and city associated with the outgoing IP.
func GetMyIPLocation() (string, error) {
	return getMyIPLocation(context.Background(), defaultHTTPClient)
}

func getMyIPLocation(ctx context.Context, client httpDoer) (string, error) {
	response := new(responseIPIP)
	if err := getJSONWithClient(ctx, client, "ipip", "https://myip.ipip.net/json", response); err != nil {
		return "", err
	}

	if response.Ret != "ok" {
		return "", fmt.Errorf("ipip: unexpected ret %q", response.Ret)
	}

	location := response.Data.Location
	if len(location) < 3 {
		return "", fmt.Errorf("ipip: data.location must contain country, province, and city; got %d entries", len(location))
	}
	for index, field := range []string{"country", "province", "city"} {
		location[index] = strings.TrimSpace(location[index])
		if location[index] == "" {
			return "", fmt.Errorf("ipip: data.location %s is empty", field)
		}
	}
	if location[0] != "中国" {
		return "", errors.New("ipip: 非境内IP地址，放弃继续解析 :(")
	}

	province, city := location[1], location[2]
	if province == "北京" || province == "上海" || province == "重庆" || province == "天津" {
		return province + "市", nil
	}
	return province + "省" + city + "市", nil
}
