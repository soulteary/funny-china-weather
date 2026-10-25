package weather_test

import (
	"fmt"
	"log"

	weather "github.com/soulteary/funny-china-weather"
)

// This example is compiled but does not execute a live provider request in tests.
func ExampleGetMyIPLocation() {
	location, err := weather.GetMyIPLocation()
	if err != nil {
		log.Print(err)
		return
	}
	fmt.Println(location)
}

// This example is compiled but does not execute a live provider request in tests.
func ExampleGetWeatherByLocation() {
	code, degree, humidity, updated, err := weather.GetWeatherByLocation("北京市")
	if err != nil {
		log.Print(err)
		return
	}
	fmt.Printf("code=%d temperature=%d°C humidity=%d%% updated=%s\n", code, degree, humidity, updated)
}

func ExampleGetWeatherIconByCode() {
	key, name := weather.GetWeatherIconByCode(-1)
	fmt.Println(key, name)
	// Output: day 未知
}
