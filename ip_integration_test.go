//go:build integration
// +build integration

package weather

import (
	"context"
	"testing"
)

func TestGetMyIPLocationIntegration(t *testing.T) {
	location, err := GetMyIPLocation()
	if err != nil {
		t.Fatalf("IPIP request failed: %v", err)
	}
	if location == "" {
		t.Fatal("IPIP returned an empty location")
	}
}

// TestIPLocationToWeatherIntegration checks that the real IPIP result is
// accepted by the weather provider without application-side normalization.
func TestIPLocationToWeatherIntegration(t *testing.T) {
	location, err := GetMyIPLocation()
	if err != nil {
		t.Fatalf("IPIP request failed: %v", err)
	}
	observation, err := GetObservation(context.Background(), location)
	if err != nil {
		t.Fatalf("weather request for automatically detected location failed: %v", err)
	}
	if _, err := ParseUpdateTime(observation.UpdateTime); err != nil {
		t.Fatalf("invalid live observation timestamp: %v", err)
	}
	t.Log("IPIP location accepted by Tencent; observation and timestamp validated")
}
