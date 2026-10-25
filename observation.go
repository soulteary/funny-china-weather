package weather

import (
	"context"
	"fmt"
	"time"
)

// Observation is a validated Tencent weather observation. Description,
// ShortDescription, and UpdateTime retain the provider's original strings.
// Obtaining an observation does not parse its timestamp or enforce freshness.
type Observation struct {
	// Code is the provider's weather code; unknown integer codes are retained.
	Code int
	// TemperatureC is the observed temperature in degrees Celsius.
	TemperatureC int
	// HumidityPercent is relative humidity in the inclusive range 0 to 100.
	HumidityPercent int
	// Description retains the original weather field without an icon-name fallback.
	Description string
	// ShortDescription retains the original weather_short field.
	ShortDescription string
	// UpdateTime retains the raw update_time field, without automatic date validation.
	UpdateTime string
}

// GetObservation returns current weather and the provider's weather descriptions.
// The request respects ctx and the shared three-second client timeout.
// ctx must not be nil. An error returns a zero Observation.
func GetObservation(ctx context.Context, location string) (Observation, error) {
	return getObservation(ctx, location, defaultHTTPClient)
}

var observationTimeZone = time.FixedZone("UTC+8", 8*60*60)

// ParseUpdateTime parses exactly 12 ASCII digits in YYYYMMDDHHMM format using
// a fixed UTC+8 offset. It does not trim or modify the input and rejects invalid
// calendar dates. It performs no freshness check; fetching weather preserves the
// raw UpdateTime even when this optional parsing step would fail.
func ParseUpdateTime(raw string) (time.Time, error) {
	if len(raw) != 12 {
		return time.Time{}, &Error{Provider: "tencent", Stage: "input", Field: "update_time", Err: fmt.Errorf("update_time must contain exactly 12 digits in YYYYMMDDHHMM format")}
	}
	for _, char := range raw {
		if char < '0' || char > '9' {
			return time.Time{}, &Error{Provider: "tencent", Stage: "input", Field: "update_time", Err: fmt.Errorf("update_time must contain exactly 12 ASCII digits")}
		}
	}
	parsed, err := time.ParseInLocation("200601021504", raw, observationTimeZone)
	if err != nil {
		return time.Time{}, &Error{Provider: "tencent", Stage: "input", Field: "update_time", Err: fmt.Errorf("invalid update_time: %w", err)}
	}
	return parsed, nil
}

// Age returns the elapsed time since the parsed UpdateTime at the supplied
// reference instant. Invalid timestamps and timestamps after at return an error
// and a zero duration. The calculation compares instants, regardless of at's zone.
func (observation Observation) Age(at time.Time) (time.Duration, error) {
	updated, err := ParseUpdateTime(observation.UpdateTime)
	if err != nil {
		return 0, err
	}
	if updated.After(at) {
		return 0, &Error{Provider: "tencent", Stage: "freshness", Field: "update_time", Err: fmt.Errorf("update_time is after the reference time")}
	}
	return at.Sub(updated), nil
}

// IsStale reports whether the observation is strictly older than maxAge at at.
// maxAge must be positive; equality is not stale. Invalid thresholds or timestamps
// and future timestamps return false and an error. No default threshold is applied.
func (observation Observation) IsStale(at time.Time, maxAge time.Duration) (bool, error) {
	if maxAge <= 0 {
		return false, &Error{Provider: "tencent", Stage: "input", Field: "max_age", Err: fmt.Errorf("max_age must be positive")}
	}
	age, err := observation.Age(at)
	if err != nil {
		return false, err
	}
	return age > maxAge, nil
}
