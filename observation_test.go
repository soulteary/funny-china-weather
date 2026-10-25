package weather

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestGetObservationPreservesProviderFields(t *testing.T) {
	fields := map[string]string{"weather_code": "999", "degree": "-4", "weather": " 上游天气描述 ", "weather_short": "上游简述", "update_time": "not-a-date"}
	got, err := getObservation(context.Background(), "北京市", fixtureClient(weatherFixture(fields)))
	want := Observation{Code: 999, TemperatureC: -4, HumidityPercent: 81, Description: fields["weather"], ShortDescription: fields["weather_short"], UpdateTime: fields["update_time"]}
	if err != nil || got != want {
		t.Fatalf("getObservation() = %#v, %v; want %#v", got, err, want)
	}
	code, degree, humidity, updated, err := getWeatherByLocation(context.Background(), "北京市", fixtureClient(weatherFixture(fields)))
	if err != nil || code != got.Code || degree != got.TemperatureC || humidity != got.HumidityPercent || updated != got.UpdateTime {
		t.Fatalf("tuple wrapper disagrees with structured observation: (%d,%d,%d,%q), %v", code, degree, humidity, updated, err)
	}
}

func TestGetObservationDoesNotEnforceFreshness(t *testing.T) {
	for _, raw := range []string{"not-a-date", "197001010000", "209901010000"} {
		t.Run(raw, func(t *testing.T) {
			got, err := getObservation(context.Background(), "北京市", fixtureClient(weatherFixture(map[string]string{"update_time": raw})))
			if err != nil || got.UpdateTime != raw {
				t.Fatalf("getObservation() = %#v, %v; want preserved raw timestamp", got, err)
			}
		})
	}
}

func TestGetObservationAllowsMissingDescriptions(t *testing.T) {
	got, err := getObservation(context.Background(), "北京市", fixtureClient(weatherFixture(nil)))
	if err != nil || got.Description != "" || got.ShortDescription != "" {
		t.Fatalf("optional descriptions = %q / %q, error = %v", got.Description, got.ShortDescription, err)
	}
}

func TestGetObservationReturnsZeroOnError(t *testing.T) {
	for _, body := range []string{`{"status":500,"message":"failed"}`, weatherFixture(map[string]string{"humidity": "101"}), weatherFixture(map[string]string{"update_time": ""})} {
		got, err := getObservation(context.Background(), "北京市", fixtureClient(body))
		if err == nil || got != (Observation{}) {
			t.Fatalf("getObservation() = %#v, %v; want zero observation and error", got, err)
		}
	}
}

func TestParseUpdateTime(t *testing.T) {
	got, err := ParseUpdateTime("202610260800")
	want := time.Date(2026, time.October, 26, 0, 0, 0, 0, time.UTC)
	if err != nil || !got.Equal(want) {
		t.Fatalf("ParseUpdateTime() = %v, %v; want %v", got, err, want)
	}
	_, offset := got.Zone()
	if offset != 8*60*60 {
		t.Fatalf("zone offset = %d, want UTC+8", offset)
	}
	if _, err := ParseUpdateTime("202402290001"); err != nil {
		t.Fatalf("valid leap day rejected: %v", err)
	}
	for _, raw := range []string{"", "20261026080", "2026102608000", "202602290001", "202613010000", "202604310001", "202610262400", "202610260060", " 202610260800", "202610260800 ", "20261026080a", "２０２６１０２６０８００"} {
		t.Run(raw, func(t *testing.T) {
			got, err := ParseUpdateTime(raw)
			var detail *Error
			if err == nil || !got.IsZero() || !errors.As(err, &detail) || detail.Stage != "input" || detail.Field != "update_time" {
				t.Fatalf("ParseUpdateTime(%q) = %v, %v; want zero time and typed input error", raw, got, err)
			}
		})
	}
	_, err = ParseUpdateTime("202602290001")
	var cause *time.ParseError
	if !errors.As(err, &cause) {
		t.Fatalf("error = %v, want wrapped calendar parse error", err)
	}
}

func TestObservationAge(t *testing.T) {
	observation := Observation{UpdateTime: "202610260800"}
	updated := time.Date(2026, time.October, 26, 0, 0, 0, 0, time.UTC)
	for _, test := range []struct {
		name      string
		at        time.Time
		want      time.Duration
		wantStage string
	}{
		{name: "same instant", at: updated},
		{name: "elapsed", at: updated.Add(30 * time.Minute), want: 30 * time.Minute},
		{name: "different zone", at: updated.Add(30 * time.Minute).In(time.FixedZone("UTC-5", -5*60*60)), want: 30 * time.Minute},
		{name: "future timestamp", at: updated.Add(-time.Nanosecond), wantStage: "freshness"},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := observation.Age(test.at)
			if test.wantStage != "" {
				var detail *Error
				if err == nil || got != 0 || !errors.As(err, &detail) || detail.Stage != test.wantStage {
					t.Fatalf("Age() = %v, %v; want zero age and %s error", got, err, test.wantStage)
				}
			} else if err != nil || got != test.want {
				t.Fatalf("Age() = %v, %v; want %v", got, err, test.want)
			}
		})
	}
	if got, err := (Observation{UpdateTime: "invalid"}).Age(updated); err == nil || got != 0 {
		t.Fatalf("invalid timestamp age = %v, %v", got, err)
	}
}

func TestObservationIsStale(t *testing.T) {
	observation := Observation{UpdateTime: "202610260800"}
	at := time.Date(2026, time.October, 26, 0, 30, 0, 0, time.UTC)
	for _, test := range []struct {
		name      string
		maxAge    time.Duration
		want      bool
		wantError bool
	}{
		{name: "older than threshold", maxAge: 29 * time.Minute, want: true},
		{name: "equal to threshold", maxAge: 30 * time.Minute},
		{name: "younger than threshold", maxAge: 31 * time.Minute},
		{name: "zero threshold", wantError: true},
		{name: "negative threshold", maxAge: -time.Minute, wantError: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := observation.IsStale(at, test.maxAge)
			if got != test.want || (err != nil) != test.wantError {
				t.Fatalf("IsStale() = %v, %v; want %v, error %v", got, err, test.want, test.wantError)
			}
			if test.wantError {
				var detail *Error
				if !errors.As(err, &detail) || detail.Stage != "input" || detail.Field != "max_age" {
					t.Fatalf("threshold error = %v", err)
				}
			}
		})
	}
	for _, raw := range []string{"invalid", "209901010000"} {
		got, err := (Observation{UpdateTime: raw}).IsStale(at, time.Hour)
		if got || err == nil {
			t.Fatalf("IsStale(%q) = %v, %v; want false and error", raw, got, err)
		}
	}
}

func TestGetObservationInFlightCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	started := make(chan struct{})
	release := make(chan struct{})
	defer close(release)
	completed := make(chan error, 1)
	client := &http.Client{Transport: fixtureTransport(func(request *http.Request) (*http.Response, error) {
		close(started)
		select {
		case <-request.Context().Done():
			return nil, request.Context().Err()
		case <-release:
			return nil, errors.New("test request released")
		}
	})}
	go func() { _, err := getObservation(ctx, "北京市", client); completed <- err }()
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("request did not start")
	}
	cancel()
	select {
	case err := <-completed:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("error = %v, want canceled request", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("request ignored cancellation")
	}
}

func TestObservationIsStaleBeyondDurationRange(t *testing.T) {
	observation := Observation{UpdateTime: "000101010000"}
	at := time.Date(2026, time.October, 26, 0, 30, 0, 0, time.UTC)
	maxAge := time.Duration(1<<63 - 1)
	age, err := observation.Age(at)
	if err != nil || age != maxAge {
		t.Fatalf("Age() = %v, %v; want saturated duration %v", age, err, maxAge)
	}
	stale, err := observation.IsStale(at, maxAge)
	if err != nil || !stale {
		t.Fatalf("IsStale() = %v, %v; ancient observation must be older than maximum duration", stale, err)
	}
}
