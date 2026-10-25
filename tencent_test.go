package weather

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
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
	code, degree, humidity, updated, err := getWeatherByLocation(context.Background(), "北京市", fixtureClient(weatherFixture(nil)))
	if err != nil {
		t.Fatal(err)
	}
	if code != 3 || degree != -4 || humidity != 81 || updated != "202610250745" {
		t.Fatalf("getWeatherByLocation() = (%d, %d, %d, %q), want (3, -4, 81, %q)", code, degree, humidity, updated, "202610250745")
	}
}

func TestGetWeatherPreservesUpdateTime(t *testing.T) {
	want := "upstream-time-unchanged"
	_, _, _, got, err := getWeatherByLocation(context.Background(), "北京市", fixtureClient(weatherFixture(map[string]string{"update_time": want})))
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
				code, degree, humidity, updated, err := getWeatherByLocation(context.Background(), "北京市", fixtureClient(weatherFixture(map[string]string{field: value})))
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
	for _, body := range []string{`{`, `{}`, `{"data":{}}`, `{"data":{"observe":null}}`, `{"status":200}`, `{"status":200,"data":{"observe":null}}`} {
		t.Run(body, func(t *testing.T) {
			_, _, _, _, err := getWeatherByLocation(context.Background(), "北京市", fixtureClient(body))
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
	for _, location := range []string{"", "广东省", "invalid"} {
		_, _, _, _, err := getWeatherByLocation(context.Background(), location, client)
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
	_, _, _, _, err := getWeatherByLocation(context.Background(), "北京市", client)
	if !errors.Is(err, requestError) {
		t.Fatalf("error = %v, want request error", err)
	}
}

func TestGetWeatherBusinessStatus(t *testing.T) {
	for _, status := range []interface{}{500, 0, nil, "missing"} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			var payload map[string]interface{}
			if err := json.Unmarshal([]byte(weatherFixture(nil)), &payload); err != nil {
				t.Fatal(err)
			}
			if status == "missing" {
				delete(payload, "status")
			} else {
				payload["status"] = status
			}
			body, err := json.Marshal(payload)
			if err != nil {
				t.Fatal(err)
			}
			code, degree, humidity, updated, err := getWeatherByLocation(context.Background(), "北京市", fixtureClient(string(body)))
			if err == nil || !strings.Contains(err.Error(), "tencent: unexpected status") || code != 0 || degree != 0 || humidity != 0 || updated != "" {
				t.Fatalf("weather = (%d, %d, %d, %q), error = %v; want zero values and status error", code, degree, humidity, updated, err)
			}
		})
	}
}

func TestGetWeatherNumericCauses(t *testing.T) {
	for _, field := range []string{"weather_code", "degree", "humidity"} {
		for _, test := range []struct {
			name, value string
			cause       error
		}{
			{name: "syntax", value: "invalid", cause: strconv.ErrSyntax},
			{name: "overflow", value: strings.Repeat("9", 100), cause: strconv.ErrRange},
		} {
			t.Run(field+"/"+test.name, func(t *testing.T) {
				_, _, _, _, err := getWeatherByLocation(context.Background(), "北京市", fixtureClient(weatherFixture(map[string]string{field: test.value})))
				if !errors.Is(err, test.cause) || !strings.Contains(err.Error(), "tencent: invalid "+field) {
					t.Fatalf("error = %v, want %s and cause %v", err, field, test.cause)
				}
			})
		}
	}
}

func TestGetWeatherHumidityAndUpdateTime(t *testing.T) {
	tests := []struct {
		name      string
		fields    map[string]string
		wantError bool
	}{
		{name: "zero humidity", fields: map[string]string{"humidity": "0"}},
		{name: "maximum humidity", fields: map[string]string{"humidity": "100"}},
		{name: "negative humidity", fields: map[string]string{"humidity": "-1"}, wantError: true},
		{name: "humidity above maximum", fields: map[string]string{"humidity": "101"}, wantError: true},
		{name: "blank update time", fields: map[string]string{"update_time": " \t"}, wantError: true},
		{name: "update time preserved", fields: map[string]string{"update_time": " upstream-time-unchanged "}},
		{name: "unknown code retained", fields: map[string]string{"weather_code": "999"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, _, _, updated, err := getWeatherByLocation(context.Background(), "北京市", fixtureClient(weatherFixture(test.fields)))
			if (err != nil) != test.wantError {
				t.Fatalf("error = %v, wantError %v", err, test.wantError)
			}
			if !test.wantError && test.fields["update_time"] != "" && updated != test.fields["update_time"] {
				t.Fatalf("update time = %q, want original %q", updated, test.fields["update_time"])
			}
		})
	}
}

func TestGetWeatherMissingFields(t *testing.T) {
	for _, field := range []string{"weather_code", "degree", "humidity", "update_time"} {
		t.Run(field, func(t *testing.T) {
			var payload map[string]interface{}
			if err := json.Unmarshal([]byte(weatherFixture(nil)), &payload); err != nil {
				t.Fatal(err)
			}
			observe := payload["data"].(map[string]interface{})["observe"].(map[string]interface{})
			delete(observe, field)
			body, err := json.Marshal(payload)
			if err != nil {
				t.Fatal(err)
			}
			_, _, _, _, err = getWeatherByLocation(context.Background(), "北京市", fixtureClient(string(body)))
			if err == nil || !strings.Contains(err.Error(), "tencent:") || !strings.Contains(err.Error(), field) {
				t.Fatalf("error = %v, want missing %s error", err, field)
			}
		})
	}
}
