package weather

import (
	"errors"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type itemTencentWeather struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Code  int    `json:"code"`
}

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

var tencentWeatherInfoMap = [...]itemTencentWeather{
	{Name: "晴", Value: "day", Code: 0},
	{Name: "晴", Value: "night", Code: 0},
	{Name: "多云", Value: "cloudy-day", Code: 1},
	{Name: "夜间多云", Value: "cloudy-night", Code: 1},
	{Name: "阴天", Value: "cloudy", Code: 2},
	{Name: "白天阵雨", Value: "rainy-1", Code: 3},
	{Name: "夜间阵雨", Value: "rainy-night-1", Code: 3},
	{Name: "打雷", Value: "thunder", Code: 4},
	{Name: "雷雨天", Value: "thunder-rain", Code: 5},
	{Name: "雨夹雪", Value: "rain-snow-day", Code: 6},
	{Name: "雨夹雪", Value: "rain-snow-night", Code: 6},
	{Name: "小雨", Value: "rainy-4", Code: 7},
	{Name: "中雨", Value: "rainy-5", Code: 8},
	{Name: "大雨", Value: "rainy-6", Code: 9},
	{Name: "雨", Value: "rainy-6", Code: 10},
	{Name: "雨", Value: "rainy-6", Code: 11},
	{Name: "雨", Value: "rainy-6", Code: 12},
	{Name: "雪天", Value: "snowy-day", Code: 13},
	{Name: "雪天", Value: "snowy-night", Code: 13},
	{Name: "小雪", Value: "snowy-4", Code: 14},
	{Name: "中雪", Value: "snowy-5", Code: 15},
	{Name: "大雪", Value: "snowy-6", Code: 16},
	{Name: "暴雪", Value: "snowy-6", Code: 17},
	{Name: "中雨", Value: "rainy-6", Code: 19},
	{Name: "中雨", Value: "rainy-6", Code: 21},
	{Name: "中雨", Value: "rainy-6", Code: 22},
	{Name: "中雨", Value: "rainy-6", Code: 23},
	{Name: "中雨", Value: "rainy-6", Code: 24},
	{Name: "中雨", Value: "rainy-6", Code: 25},
	{Name: "中雪", Value: "snowy-6", Code: 26},
	{Name: "中雪", Value: "snowy-6", Code: 27},
	{Name: "中雪", Value: "snowy-6", Code: 28},
	{Name: "有雾", Value: "icon-fog", Code: 18},
	{Name: "有雾", Value: "icon-fog", Code: 32},
	{Name: "沙尘暴", Value: "icon-dust", Code: 20},
	{Name: "沙尘暴", Value: "icon-dust", Code: 29},
	{Name: "沙尘暴", Value: "icon-dust", Code: 30},
	{Name: "沙尘暴", Value: "icon-dust", Code: 31},
	{Name: "冰雹", Value: "icon-hail", Code: -1},
	{Name: "霰", Value: "icon-sleet", Code: -1},
	{Name: "风", Value: "icon-wind", Code: -1},
}

// 解析来自腾讯天气的接口数据（中国气象网）
func GetWeatherByLocation(location string) (code int, degree int, humidity int, updateTime string, err error) {

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

	fetchErr := getJSON("https://wis.qq.com/weather/common?source=pc&weather_type=observe&province="+province+"&city="+city, 3, response)
	if fetchErr != nil {
		return 0, 0, 0, "", fetchErr
	}

	code, parseIntError := strconv.Atoi(response.Data.Observe.WeatherCode)
	if parseIntError != nil {
		return 0, 0, 0, "", parseIntError
	}
	degree, parseIntError = strconv.Atoi(response.Data.Observe.Degree)
	if parseIntError != nil {
		return 0, 0, 0, "", parseIntError
	}

	humidity, parseIntError = strconv.Atoi(response.Data.Observe.Humidity)
	if parseIntError != nil {
		return 0, 0, 0, "", parseIntError
	}

	return code, degree, humidity, updateTime, nil
}

func GetWeatherIconByCode(weatherCode int) (string, string) {
	hour, _, _ := time.Now().Clock()
	isDay := hour >= 5 && hour <= 18

	var codeMatched []itemTencentWeather
	for idx := range tencentWeatherInfoMap {
		if tencentWeatherInfoMap[idx].Code == weatherCode {
			codeMatched = append(codeMatched, tencentWeatherInfoMap[idx])
		}
	}

	if len(codeMatched) > 1 {
		if isDay {
			for idx := range codeMatched {
				if strings.HasSuffix(codeMatched[idx].Value, "day") {
					return codeMatched[idx].Value, codeMatched[idx].Name
				}
			}
		} else {
			for idx := range codeMatched {
				if strings.HasSuffix(codeMatched[idx].Value, "night") {
					return codeMatched[idx].Value, codeMatched[idx].Name
				}
			}
		}
	} else if len(codeMatched) == 1 {
		return codeMatched[0].Value, codeMatched[0].Name
	}

	return "day", "未知"
}
