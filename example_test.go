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

func ExampleGetWeatherIconByCode() {
	key, name := weather.GetWeatherIconByCode(-1)
	fmt.Println(key, name)
	// Output: unknown 未知
}
