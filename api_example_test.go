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

func ExampleGetWeatherByLocationContext() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, _, _, _, err := weather.GetWeatherByLocationContext(ctx, "北京市")
	fmt.Println(errors.Is(err, context.Canceled))
	// Output: true
}

func ExampleGetObservation() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := weather.GetObservation(ctx, "北京市")
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
