package weather

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
)

func weatherFixture(overrides map[string]string) string {
	observe := map[string]string{
		"weather_code": "3",
		"degree":       "-4",
		"humidity":     "81",
		"update_time":  "202610250745",
	}
	for key, value := range overrides {
		observe[key] = value
	}
	body, _ := json.Marshal(map[string]interface{}{
		"status":  200,
		"message": "OK",
		"data":    map[string]interface{}{"observe": observe},
	})
	return string(body)
}

func TestGetWeatherByLocation(t *testing.T) {
	code, degree, humidity, updated, err := getWeatherByLocation("北京市", fixtureClient(weatherFixture(nil)))
	if err != nil {
		t.Fatal(err)
	}
	if code != 3 || degree != -4 || humidity != 81 || updated != "202610250745" {
		t.Fatalf("getWeatherByLocation() = (%d, %d, %d, %q), want (3, -4, 81, %q)", code, degree, humidity, updated, "202610250745")
	}
}

func TestGetWeatherPreservesUpdateTime(t *testing.T) {
	want := "upstream-time-unchanged"
	_, _, _, got, err := getWeatherByLocation("北京市", fixtureClient(weatherFixture(map[string]string{"update_time": want})))
	if err != nil || got != want {
		t.Fatalf("update time = %q, error = %v; want %q without an error", got, err, want)
	}
}

func TestGetWeatherRequiredFields(t *testing.T) {
	for _, field := range []string{"weather_code", "degree", "humidity", "update_time"} {
		for _, value := range []string{"", "invalid"} {
			if field == "update_time" && value != "" {
				continue
			}
			t.Run(field+"/"+value, func(t *testing.T) {
				code, degree, humidity, updated, err := getWeatherByLocation("北京市", fixtureClient(weatherFixture(map[string]string{field: value})))
				if err == nil || !strings.Contains(err.Error(), field) {
					t.Fatalf("error = %v, want an error identifying %s", err, field)
				}
				if code != 0 || degree != 0 || humidity != 0 || updated != "" {
					t.Fatalf("failed response returned weather values: (%d, %d, %d, %q)", code, degree, humidity, updated)
				}
			})
		}
	}
}

func TestGetWeatherInvalidPayload(t *testing.T) {
	for _, body := range []string{`{`, `{}`, `{"data":{}}`, `{"data":{"observe":null}}`} {
		t.Run(body, func(t *testing.T) {
			_, _, _, _, err := getWeatherByLocation("北京市", fixtureClient(body))
			if err == nil {
				t.Fatal("expected malformed or incomplete weather data to fail")
			}
		})
	}
}

func TestGetWeatherRejectsInvalidLocationBeforeRequest(t *testing.T) {
	client := &http.Client{Transport: fixtureTransport(func(*http.Request) (*http.Response, error) {
		t.Fatal("invalid location triggered an HTTP request")
		return nil, nil
	})}
	for _, location := range []string{"", "北京", "invalid"} {
		_, _, _, _, err := getWeatherByLocation(location, client)
		if err == nil {
			t.Errorf("location %q should fail", location)
		}
	}
}

func TestGetWeatherRequestFailure(t *testing.T) {
	requestError := errors.New("request failed")
	client := &http.Client{Transport: fixtureTransport(func(*http.Request) (*http.Response, error) {
		return nil, requestError
	})}
	_, _, _, _, err := getWeatherByLocation("北京市", client)
	if !errors.Is(err, requestError) {
		t.Fatalf("error = %v, want request error", err)
	}
}
