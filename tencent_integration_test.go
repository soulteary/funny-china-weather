//go:build integration
// +build integration

package weather

import "testing"

func TestGetWeatherByLocationIntegration(t *testing.T) {
	_, _, _, updateTime, err := GetWeatherByLocation("北京市")
	if err != nil {
		t.Fatalf("Tencent request failed: %v", err)
	}
	if updateTime == "" {
		t.Fatal("Tencent returned an empty update time")
	}
}
