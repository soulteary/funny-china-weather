//go:build integration
// +build integration

package weather

import "testing"

func TestGetMyIPLocationIntegration(t *testing.T) {
	location, err := GetMyIPLocation()
	if err != nil {
		t.Fatalf("IPIP request failed: %v", err)
	}
	if location == "" {
		t.Fatal("IPIP returned an empty location")
	}
}
