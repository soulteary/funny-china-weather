package flare

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	weather "github.com/soulteary/funny-china-weather"
)

func TestRefreshDoesNotBlockOtherLocations(t *testing.T) {
	c := NewCache()
	started, release := make(chan struct{}), make(chan struct{})
	finished := make(chan struct{})
	defer func() { close(release); <-finished }()
	var calls atomic.Int32
	c.fetch = func(ctx context.Context, key string) (weather.Observation, error) {
		calls.Add(1)
		if key == "slow" {
			close(started)
			<-release
		}
		return weather.Observation{UpdateTime: "202610260800"}, nil
	}
	c.Get(context.Background(), "cached")
	go func() { c.Get(context.Background(), "slow"); close(finished) }()
	<-started
	for _, key := range []string{"cached", "other"} {
		done := make(chan struct{})
		go func() { c.Get(context.Background(), key); close(done) }()
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("unrelated location blocked")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	v := c.Get(ctx, "slow")
	if !errors.Is(v.LastError, context.Canceled) {
		t.Fatalf("cancellation: %v", v.LastError)
	}
	if calls.Load() != 3 {
		t.Fatalf("unexpected requests: %d", calls.Load())
	}
	// The cleanup releases the original refresh even if this test fails.
}

func TestWaitingRefreshCanBeCanceled(t *testing.T) {
	c := NewCache()
	started, release := make(chan struct{}), make(chan struct{})
	c.fetch = func(context.Context, string) (weather.Observation, error) {
		close(started)
		<-release
		return weather.Observation{}, nil
	}
	done := make(chan struct{})
	go func() { c.Get(context.Background(), "same"); close(done) }()
	<-started
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	v := c.Get(ctx, "same")
	close(release)
	<-done
	if !errors.Is(v.LastError, context.DeadlineExceeded) {
		t.Fatalf("wait cancellation: %v", v.LastError)
	}
}

func TestConcurrentRefreshIsCoalesced(t *testing.T) {
	c := NewCache()
	started, release := make(chan struct{}), make(chan struct{})
	var calls atomic.Int32
	c.fetch = func(context.Context, string) (weather.Observation, error) {
		if calls.Add(1) == 1 {
			close(started)
		}
		<-release
		return weather.Observation{}, nil
	}
	done := make(chan View, 20)
	for i := 0; i < 20; i++ {
		go func() { done <- c.Get(context.Background(), "same") }()
	}
	<-started
	close(release)
	for i := 0; i < 20; i++ {
		select {
		case v := <-done:
			if !v.Available {
				t.Fatal("missing shared result")
			}
		case <-time.After(time.Second):
			t.Fatal("waiter blocked")
		}
	}
	if calls.Load() != 1 {
		t.Fatalf("duplicate refreshes: %d", calls.Load())
	}
}

func TestNilContextDoesNotFetch(t *testing.T) {
	c := NewCache()
	c.fetch = func(context.Context, string) (weather.Observation, error) {
		t.Fatal("nil context started a request")
		return weather.Observation{}, nil
	}
	for _, v := range []View{c.Get(nil, "北京"), c.GetAt(nil, weather.Location{Province: "北京", City: "北京"})} {
		var err *weather.Error
		if v.Available || !v.Stale || !errors.As(v.LastError, &err) || err.Stage != "input" || err.Field != "context" {
			t.Fatalf("nil context validation: %+v", v)
		}
	}
}

func TestRefreshCancellationPreservesDataAndRetry(t *testing.T) {
	c := NewCache()
	at := time.Date(2026, 10, 26, 0, 0, 0, 0, time.UTC)
	c.now = func() time.Time { return at }
	previous := weather.Observation{TemperatureC: 20, UpdateTime: "202610260800"}
	c.fetch = func(context.Context, string) (weather.Observation, error) { return previous, nil }
	c.Get(context.Background(), "same")
	at = at.Add(refreshInterval)
	started := make(chan struct{})
	c.fetch = func(ctx context.Context, _ string) (weather.Observation, error) {
		close(started)
		<-ctx.Done()
		return weather.Observation{}, ctx.Err()
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan View, 1)
	go func() { done <- c.Get(ctx, "same") }()
	<-started
	cancel()
	v := <-done
	if !v.Available || !v.Stale || v.Observation != previous || !errors.Is(v.LastError, context.Canceled) || v.NextAttempt != at.Add(retryInterval) {
		t.Fatalf("canceled refresh: %+v", v)
	}
	// A subsequent caller must use the retained entry, not start another request.
	v = c.Get(context.Background(), "same")
	if !v.Available || !v.Stale || v.Observation != previous {
		t.Fatalf("retry cache: %+v", v)
	}
}
