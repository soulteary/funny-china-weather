package weather

import "time"

// GetWeatherIconAt selects an icon using the clock and time zone carried by at.
// Daytime starts at 05:00 and ends immediately before 19:00. Unknown weather
// codes return "day" and "未知", including at night.
func GetWeatherIconAt(weatherCode int, at time.Time) (key, name string) {
	return iconByCodeAt(weatherCode, at)
}
