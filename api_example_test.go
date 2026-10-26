package weather_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	weather "github.com/soulteary/funny-china-weather"
)

func ExampleGetMyIPLocationContext() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // A canceled context stops the request before any network access.
	_, err := weather.GetMyIPLocationContext(ctx)
	fmt.Println(errors.Is(err, context.Canceled))
	// Output: true
}

func ExampleGetObservationAt_cancellation() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := weather.GetObservationAt(ctx, weather.Location{Province: "北京", City: "北京"})
	var detail *weather.Error
	if errors.As(err, &detail) {
		fmt.Println(detail.Provider, detail.Stage, errors.Is(err, context.Canceled))
	}
	// Output: tencent request true
}

func ExampleObservation() {
	observation := weather.Observation{
		Code: 301, TemperatureC: 20, HumidityPercent: 80,
		Description: "雨", ShortDescription: "雨", UpdateTime: "202610260800",
	}
	fmt.Println(observation.Code, observation.Description)
	// Output: 301 雨
}

func ExampleParseUpdateTime() {
	updated, err := weather.ParseUpdateTime("202610260800")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(updated.UTC().Format(time.RFC3339))
	// Output: 2026-10-26T00:00:00Z
}

func ExampleObservation_Age() {
	observation := weather.Observation{UpdateTime: "202610260800"}
	at := time.Date(2026, time.October, 26, 0, 30, 0, 0, time.UTC)
	age, err := observation.Age(at)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(age)
	// Output: 30m0s
}

func ExampleObservation_IsStale() {
	observation := weather.Observation{UpdateTime: "202610260800"}
	at := time.Date(2026, time.October, 26, 0, 30, 0, 0, time.UTC)
	stale, err := observation.IsStale(at, 15*time.Minute)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(stale)
	// Output: true
}

func ExampleError() {
	err := &weather.Error{Provider: "tencent", Stage: "http", HTTPStatus: 429}
	var detail *weather.Error
	if errors.As(err, &detail) {
		fmt.Println(detail.Provider, detail.Stage, detail.HTTPStatus)
	}
	// Output: tencent http 429
}

func ExampleErrResponseTooLarge() {
	err := &weather.Error{Provider: "tencent", Stage: "size", Err: weather.ErrResponseTooLarge}
	fmt.Println(errors.Is(err, weather.ErrResponseTooLarge))
	// Output: true
}

// Search once, let the caller choose a candidate, then reuse the location.
// This example is compiled but does not make live requests during normal tests.
func ExampleSearchLocations() {
	locations, err := weather.SearchLocations(context.Background(), "朝阳")
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, location := range locations {
		fmt.Println(location.ID, location.Province, location.City, location.County)
	}
}

func ExampleGetObservationAt() {
	locations, err := weather.SearchLocations(context.Background(), "津市")
	if err != nil {
		fmt.Println(err)
		return
	}
	if len(locations) != 1 {
		fmt.Println("select a location first")
		return
	}
	observation, err := weather.GetObservationAt(context.Background(), locations[0])
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(observation.TemperatureC)
}
