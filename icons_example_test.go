package weather_test

import (
	"fmt"
	"strings"

	weather "github.com/soulteary/funny-china-weather"
)

func ExampleGetIconSVG() {
	svg, ok := weather.GetIconSVG("cloudy-day")
	fmt.Println(ok, strings.Contains(svg, "<svg"))
	// Output: true true
}
