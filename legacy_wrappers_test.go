package weather

import "testing"

func TestLegacyProviderWrappers(t *testing.T) {
	original := defaultHTTPClient
	t.Cleanup(func() { defaultHTTPClient = original })
	defaultHTTPClient = fixtureClient(`{"ret":"ok","data":{"location":["中国","北京","北京"]}}`)
	if got, err := GetMyIPLocation(); err != nil || got != "北京市" {
		t.Fatalf("GetMyIPLocation()=%q, %v", got, err)
	}
	defaultHTTPClient = fixtureClient(weatherFixture(nil))
	code, degree, humidity, updated, err := GetWeatherByLocation("北京市")
	if err != nil {
		t.Fatal(err)
	}
	// The legacy tuple must match the values returned by the same fixture.
	expected, fixtureErr := getObservation(t.Context(), "北京市", defaultHTTPClient)
	if fixtureErr != nil {
		t.Fatal(fixtureErr)
	}
	if code != expected.Code || degree != expected.TemperatureC || humidity != expected.HumidityPercent || updated != expected.UpdateTime {
		t.Fatalf("legacy tuple differs from observation: %d %d %d %s", code, degree, humidity, updated)
	}
}
