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
	0:   {"clear-day", "clear-night", "晴", "晴"},
	1:   {"cloudy-day", "cloudy-night", "多云", "夜间多云"},
	2:   {dayKey: "overcast", dayName: "阴天"},
	3:   {"showers-day", "showers-night", "白天阵雨", "夜间阵雨"},
	4:   {dayKey: "thunderstorm", dayName: "雷阵雨"},
	5:   {dayKey: "thunderstorm-hail", dayName: "雷阵雨伴有冰雹"},
	6:   {"rain-snow-day", "rain-snow-night", "雨夹雪", "雨夹雪"},
	7:   {dayKey: "rain-light", dayName: "小雨"},
	8:   {dayKey: "rain-moderate", dayName: "中雨"},
	9:   {dayKey: "rain-heavy", dayName: "大雨"},
	10:  {dayKey: "rain-heavy", dayName: "暴雨"},
	11:  {dayKey: "rain-heavy", dayName: "大暴雨"},
	12:  {dayKey: "rain-heavy", dayName: "特大暴雨"},
	13:  {"snow-showers-day", "snow-showers-night", "阵雪", "阵雪"},
	14:  {dayKey: "snow-light", dayName: "小雪"},
	15:  {dayKey: "snow-moderate", dayName: "中雪"},
	16:  {dayKey: "snow-heavy", dayName: "大雪"},
	17:  {dayKey: "snow-heavy", dayName: "暴雪"},
	18:  {dayKey: "fog", dayName: "有雾"},
	19:  {dayKey: "freezing-rain", dayName: "冻雨"},
	20:  {dayKey: "dust", dayName: "沙尘暴"},
	21:  {dayKey: "rain-moderate", dayName: "小到中雨"},
	22:  {dayKey: "rain-heavy", dayName: "中到大雨"},
	23:  {dayKey: "rain-heavy", dayName: "大到暴雨"},
	24:  {dayKey: "rain-heavy", dayName: "暴雨到大暴雨"},
	25:  {dayKey: "rain-heavy", dayName: "大暴雨到特大暴雨"},
	26:  {dayKey: "snow-moderate", dayName: "小到中雪"},
	27:  {dayKey: "snow-heavy", dayName: "中到大雪"},
	28:  {dayKey: "snow-heavy", dayName: "大到暴雪"},
	29:  {dayKey: "dust", dayName: "浮尘"},
	30:  {dayKey: "dust", dayName: "扬沙"},
	31:  {dayKey: "dust", dayName: "强沙尘暴"},
	32:  {dayKey: "fog", dayName: "浓雾"},
	53:  {dayKey: "fog", dayName: "霾"},
	301: {dayKey: "rain-light", dayName: "雨"},
	302: {dayKey: "snow-light", dayName: "雪"},
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
		return "unknown", "未知"
	}
	if (at.Hour() < 5 || at.Hour() >= 19) && icons.nightKey != "" {
		return icons.nightKey, icons.nightName
	}
	return icons.dayKey, icons.dayName
}
