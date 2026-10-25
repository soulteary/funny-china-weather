package weather

import "time"

type weatherIconPair struct {
	dayKey    string
	nightKey  string
	dayName   string
	nightName string
}

// The available artwork covers fewer distinct conditions than the legend.
// Shared icons retain the precise description of each weather code.
var weatherIconsByCode = map[int]weatherIconPair{
	0:   {"day", "night", "晴", "晴"},
	1:   {"cloudy-day", "cloudy-night", "多云", "夜间多云"},
	2:   {dayKey: "cloudy", dayName: "阴天"},
	3:   {"rainy-1", "rainy-night-1", "白天阵雨", "夜间阵雨"},
	4:   {dayKey: "thunder-rain", dayName: "雷阵雨"},
	5:   {dayKey: "icon-hail", dayName: "雷阵雨伴有冰雹"},
	6:   {"rain-snow-day", "rain-snow-night", "雨夹雪", "雨夹雪"},
	7:   {dayKey: "rainy-4", dayName: "小雨"},
	8:   {dayKey: "rainy-5", dayName: "中雨"},
	9:   {dayKey: "rainy-6", dayName: "大雨"},
	10:  {dayKey: "rainy-6", dayName: "暴雨"},
	11:  {dayKey: "rainy-6", dayName: "大暴雨"},
	12:  {dayKey: "rainy-6", dayName: "特大暴雨"},
	13:  {"snowy-day", "snowy-night", "阵雪", "阵雪"},
	14:  {dayKey: "snowy-4", dayName: "小雪"},
	15:  {dayKey: "snowy-5", dayName: "中雪"},
	16:  {dayKey: "snowy-6", dayName: "大雪"},
	17:  {dayKey: "snowy-6", dayName: "暴雪"},
	18:  {dayKey: "icon-fog", dayName: "有雾"},
	19:  {dayKey: "icon-sleet", dayName: "冻雨"},
	20:  {dayKey: "icon-dust", dayName: "沙尘暴"},
	21:  {dayKey: "rainy-5", dayName: "小到中雨"},
	22:  {dayKey: "rainy-6", dayName: "中到大雨"},
	23:  {dayKey: "rainy-6", dayName: "大到暴雨"},
	24:  {dayKey: "rainy-6", dayName: "暴雨到大暴雨"},
	25:  {dayKey: "rainy-6", dayName: "大暴雨到特大暴雨"},
	26:  {dayKey: "snowy-5", dayName: "小到中雪"},
	27:  {dayKey: "snowy-6", dayName: "中到大雪"},
	28:  {dayKey: "snowy-6", dayName: "大到暴雪"},
	29:  {dayKey: "icon-dust", dayName: "浮尘"},
	30:  {dayKey: "icon-dust", dayName: "扬沙"},
	31:  {dayKey: "icon-dust", dayName: "强沙尘暴"},
	32:  {dayKey: "icon-fog", dayName: "浓雾"},
	53:  {dayKey: "icon-fog", dayName: "霾"},
	301: {dayKey: "rainy-4", dayName: "雨"},
	302: {dayKey: "snowy-4", dayName: "雪"},
}

// GetWeatherIconByCode selects an icon using the current local time.
// Daytime starts at 05:00 and ends immediately before 19:00.
// Codes 53 (haze), 301 (rain), and 302 (snow) follow China Weather's official
// legend at https://www.weather.com.cn/pub/legend.shtml. Supporting those
// legend codes does not imply that Tencent's observe endpoint emits them.
func GetWeatherIconByCode(weatherCode int) (string, string) {
	return iconByCodeAt(weatherCode, time.Now())
}

func iconByCodeAt(weatherCode int, at time.Time) (string, string) {
	icons, ok := weatherIconsByCode[weatherCode]
	if !ok {
		return "day", "未知"
	}
	if (at.Hour() < 5 || at.Hour() >= 19) && icons.nightKey != "" {
		return icons.nightKey, icons.nightName
	}
	return icons.dayKey, icons.dayName
}
