package weather

import (
	"testing"
)

func TestGetWeather(t *testing.T) {
	t.Run("Test `GetWeatherByLocation` API is alive", func(t *testing.T) {
		_, _, _, _, err := GetWeatherByLocation("北京市")
		if err != nil {
			t.Errorf("`wis.qq.com` error '%s'", err)
		}
	})

	t.Run("Test `GetWeatherIconByCode`", func(t *testing.T) {
		value, _ := GetWeatherIconByCode(123)
		if value != "day" {
			t.Errorf("未匹配图标，返回结果不正确")
		}

		value, _ = GetWeatherIconByCode(1)
		if value != "cloudy-day" {
			t.Errorf("多云天气，返回结果不正确")
		}

	})
}
