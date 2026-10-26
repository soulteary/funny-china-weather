package flare

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	weather "github.com/soulteary/funny-china-weather"
)

func TestCacheCapacityAndIdleEviction(t *testing.T) {
	c := NewCache()
	at := time.Date(2026, 10, 26, 0, 0, 0, 0, time.UTC)
	c.now = func() time.Time { return at }
	c.fetch = func(context.Context, string) (weather.Observation, error) { return weather.Observation{}, nil }
	for i := 0; i < maxCacheEntries; i++ {
		c.Get(context.Background(), fmt.Sprint(i))
		at = at.Add(time.Second)
	}
	c.Get(context.Background(), "0") // Keep the oldest location active.
	c.Get(context.Background(), "new")
	if len(c.entries) != maxCacheEntries {
		t.Fatalf("capacity: %d", len(c.entries))
	}
	if _, ok := c.entries["query:0"]; !ok {
		t.Fatal("recently used entry evicted")
	}
	if _, ok := c.entries["query:1"]; ok {
		t.Fatal("oldest entry retained")
	}
	at = at.Add(cacheIdleTTL)
	c.Get(context.Background(), "after-idle")
	if len(c.entries) != 1 {
		t.Fatalf("idle eviction: %d", len(c.entries))
	}
}

func TestFullCachePinsRefreshesAndWaits(t *testing.T) {
	c := NewCache()
	started := make(chan struct{}, maxCacheEntries+1)
	release := make(chan struct{})
	var calls atomic.Int32
	c.fetch = func(context.Context, string) (weather.Observation, error) {
		calls.Add(1)
		started <- struct{}{}
		<-release
		return weather.Observation{}, nil
	}
	var wg sync.WaitGroup
	var releaseOnce sync.Once
	releaseAll := func() { releaseOnce.Do(func() { close(release) }) }
	defer func() { releaseAll(); wg.Wait() }()
	for i := 0; i < maxCacheEntries; i++ {
		wg.Add(1)
		go func(i int) { defer wg.Done(); c.Get(context.Background(), fmt.Sprint(i)) }(i)
	}
	for i := 0; i < maxCacheEntries; i++ {
		<-started
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	v := c.Get(ctx, "overflow")
	if v.LastError != context.DeadlineExceeded {
		t.Fatalf("capacity wait: %v", v.LastError)
	}
	c.mu.Lock()
	if len(c.entries) != maxCacheEntries || len(c.inflight) != maxCacheEntries {
		t.Error("refreshing entries evicted")
	}
	c.mu.Unlock()
	if calls.Load() != maxCacheEntries {
		t.Fatal("overflow started another request")
	}
	finished := make(chan View, 1)
	wg.Add(1)
	go func() { defer wg.Done(); finished <- c.Get(context.Background(), "overflow") }()
	releaseAll()
	select {
	case v := <-finished:
		if !v.Available {
			t.Fatalf("overflow did not resume: %+v", v)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("capacity waiter did not resume")
	}
	wg.Wait()
	if calls.Load() != maxCacheEntries+1 {
		t.Fatalf("duplicate refresh during eviction: %d", calls.Load())
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.entries) > maxCacheEntries {
		t.Fatal("capacity exceeded after completion")
	}

}

func TestGetAtRefreshSkipsDirectory(t *testing.T) {
	original := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = original })
	calls := 0
	http.DefaultTransport = transportFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.URL.Host != "wis.qq.com" || r.URL.Query().Get("county") != "海淀" {
			t.Fatalf("unexpected request: %s", r.URL)
		}
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"status":200,"data":{"observe":{"weather_code":"0","degree":"20","humidity":"50","update_time":"202610260800"}}}`))}, nil
	})
	c := NewCache()
	at := time.Date(2026, 10, 26, 0, 0, 0, 0, time.UTC)
	c.now = func() time.Time { return at }
	selected := weather.Location{Province: "北京", City: "北京", County: "海淀"}
	for i := 0; i < 2; i++ {
		if v := c.GetAt(context.Background(), selected); !v.Available || v.Location != selected.String() {
			t.Fatalf("selected location: %+v", v)
		}
	}
	at = at.Add(refreshInterval)
	c.GetAt(context.Background(), selected)
	if calls != 2 {
		t.Fatalf("want one request per refresh, got %d", calls)
	}
}
