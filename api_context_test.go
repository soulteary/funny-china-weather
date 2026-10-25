package weather_test

import (
	"context"
	"errors"
	"testing"
	"time"

	weather "github.com/soulteary/funny-china-weather"
)

func TestPublicContextCancellation(t *testing.T) {
	for _, expired := range []bool{false, true} {
		var ctx context.Context
		var cancel context.CancelFunc
		cause := context.Canceled
		if expired {
			ctx, cancel = context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
			cause = context.DeadlineExceeded
		} else {
			ctx, cancel = context.WithCancel(context.Background())
			cancel()
		}
		for _, test := range []struct {
			name, provider string
			call           func(context.Context) error
		}{
			{name: "IP location", provider: "ipip", call: func(ctx context.Context) error {
				location, err := weather.GetMyIPLocationContext(ctx)
				if location != "" {
					t.Errorf("canceled call returned location %q", location)
				}
				return err
			}},
			{name: "weather tuple", provider: "tencent", call: func(ctx context.Context) error {
				code, degree, humidity, updated, err := weather.GetWeatherByLocationContext(ctx, "北京市")
				if code != 0 || degree != 0 || humidity != 0 || updated != "" {
					t.Error("canceled call returned partial weather")
				}
				return err
			}},
			{name: "observation", provider: "tencent", call: func(ctx context.Context) error {
				observation, err := weather.GetObservation(ctx, "北京市")
				if observation != (weather.Observation{}) {
					t.Errorf("canceled call returned observation %#v", observation)
				}
				return err
			}},
		} {
			t.Run(test.name+"/"+cause.Error(), func(t *testing.T) {
				err := test.call(ctx)
				var detail *weather.Error
				if !errors.Is(err, cause) || !errors.As(err, &detail) || detail.Provider != test.provider || detail.Stage != "request" {
					t.Fatalf("error = %v, want %s request error retaining %v", err, test.provider, cause)
				}
			})
		}
		cancel()
	}
}

func TestPublicNilContext(t *testing.T) {
	calls := []func() error{
		func() error { _, err := weather.GetMyIPLocationContext(nil); return err },
		func() error { _, _, _, _, err := weather.GetWeatherByLocationContext(nil, "北京市"); return err },
		func() error { _, err := weather.GetObservation(nil, "北京市"); return err },
	}
	for _, call := range calls {
		err := call()
		var detail *weather.Error
		if !errors.As(err, &detail) || detail.Stage != "input" || detail.Field != "context" {
			t.Fatalf("nil context error = %v, want typed context input error", err)
		}
	}
}
