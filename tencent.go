package weather

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
)

type responseTencentWeather struct {
	Data struct {
		Observe struct {
			Degree        string `json:"degree"`
			Humidity      string `json:"humidity"`
			Precipitation string `json:"precipitation"`
			Pressure      string `json:"pressure"`
			UpdateTime    string `json:"update_time"`
			Weather       string `json:"weather"`
			WeatherCode   string `json:"weather_code"`
			WeatherShort  string `json:"weather_short"`
			WindDirection string `json:"wind_direction"`
			WindPower     string `json:"wind_power"`
		} `json:"observe"`
	} `json:"data"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

// GetWeatherByLocation returns current weather and the unmodified upstream update time.
func GetWeatherByLocation(location string) (code int, degree int, humidity int, updateTime string, err error) {
	return getWeatherByLocation(location, newHTTPClient(3))
}

func getWeatherByLocation(location string, client httpDoer) (code int, degree int, humidity int, updateTime string, err error) {

	province := ""
	city := ""

	if location == "北京市" || location == "上海市" || location == "重庆市" || location == "天津市" {
		province = location
		city = province
	} else {
		r, _ := regexp.Compile(`(\S+省)(\S+市)`)
		matched := r.FindAllStringSubmatch(location, -1)

		if len(matched) > 0 && len(matched[0]) == 3 {
			province = url.QueryEscape(matched[0][1])
			city = url.QueryEscape(matched[0][2])
		} else {
			return 0, 0, 0, "", errors.New("地址位置参数内容有误")
		}
	}

	if province == "" || city == "" {
		return 0, 0, 0, "", errors.New("地址位置参数不能为空")
	}

	response := new(responseTencentWeather)

	fetchErr := getJSONWithClient(client, "https://wis.qq.com/weather/common?source=pc&weather_type=observe&province="+province+"&city="+city, response)
	if fetchErr != nil {
		return 0, 0, 0, "", fetchErr
	}

	code, parseIntError := strconv.Atoi(response.Data.Observe.WeatherCode)
	if parseIntError != nil {
		return 0, 0, 0, "", fmt.Errorf("tencent: invalid weather_code: %w", parseIntError)
	}
	degree, parseIntError = strconv.Atoi(response.Data.Observe.Degree)
	if parseIntError != nil {
		return 0, 0, 0, "", fmt.Errorf("tencent: invalid degree: %w", parseIntError)
	}

	humidity, parseIntError = strconv.Atoi(response.Data.Observe.Humidity)
	if parseIntError != nil {
		return 0, 0, 0, "", fmt.Errorf("tencent: invalid humidity: %w", parseIntError)
	}

	updateTime = response.Data.Observe.UpdateTime
	if updateTime == "" {
		return 0, 0, 0, "", errors.New("tencent: update_time is empty")
	}
	return code, degree, humidity, updateTime, nil
}
