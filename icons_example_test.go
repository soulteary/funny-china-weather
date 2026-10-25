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

func ExampleGetIconSVGWithPrefix() {
	first, _ := weather.GetIconSVGWithPrefix("cloudy-day", "forecast")
	again, _ := weather.GetIconSVGWithPrefix("cloudy-day", "forecast")
	fmt.Println(first == again)
	// Output: true
}

func ExampleGetStaticIconSVG() {
	svg, ok := weather.GetStaticIconSVG("fog")
	fmt.Println(ok, strings.Contains(svg, "<animate"))
	// Output: true false
}
