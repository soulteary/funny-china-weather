// Package flare demonstrates a Flare-style weather consumer with bounded retries.
// It is example application code; the weather library itself has no cache policy.
package flare

import (
	"context"
	"fmt"
	"html/template"
	"sync"
	"time"

	weather "github.com/soulteary/funny-china-weather"
)

const (
	refreshInterval   = 10 * time.Minute
	retryInterval     = time.Minute
	maxObservationAge = 30 * time.Minute
	maxCacheEntries   = 128
	cacheIdleTTL      = time.Hour
)

// View is the data passed to a home-page template. Hide numeric values and the
// icon when Available is false; show an expired badge when Stale is true.
type View struct {
	Location      string
	Observation   weather.Observation
	ConditionCode string
	ConditionText string
	WeatherIcon   template.HTML
	Available     bool
	Stale         bool
	NextAttempt   time.Time
	LastError     error
}

type entry struct {
	observation weather.Observation
	available   bool
	nextAttempt time.Time
	lastError   error
	lastUsed    time.Time
}

type refresh struct {
	done      chan struct{}
	result    entry
	completed time.Time
}

// Cache keeps separate data for each location and coalesces refreshes per key.
// Network requests and rendering run without holding the cache mutex.
type Cache struct {
	mu       sync.Mutex
	entries  map[string]entry
	inflight map[string]*refresh
	changed  chan struct{}
	fetch    func(context.Context, string) (weather.Observation, error)
	now      func() time.Time
}

// NewCache uses the current project's provider API. Render icons in the
// process's time zone, matching Flare's existing behavior.
func NewCache() *Cache {
	return &Cache{entries: make(map[string]entry), fetch: weather.GetObservation, now: time.Now}
}

// Get refreshes at most once per ten minutes after success, or once per minute
// after failure. Failed refreshes preserve the last successful observation and
// mark it stale. No data from another location is used as a fallback.
func (c *Cache) Get(ctx context.Context, location string) View {
	return c.get(ctx, "query:"+location, location, func(ctx context.Context) (weather.Observation, error) { return c.fetch(ctx, location) })
}

// GetAt refreshes a previously selected provider location without directory searches.
func (c *Cache) GetAt(ctx context.Context, location weather.Location) View {
	key := fmt.Sprintf("at:%q|%q|%q", location.Province, location.City, location.County)
	return c.get(ctx, key, location.String(), func(ctx context.Context) (weather.Observation, error) { return weather.GetObservationAt(ctx, location) })
}

func (c *Cache) get(ctx context.Context, location, display string, fetch func(context.Context) (weather.Observation, error)) View {
	if ctx == nil {
		return View{Location: display, Stale: true, LastError: &weather.Error{Provider: "tencent", Stage: "input", Field: "context", Err: fmt.Errorf("context must not be nil")}}
	}
	for {
		c.mu.Lock()
		at := c.now()
		c.evictIdle(at)
		e := c.entries[location]
		if err := ctx.Err(); err != nil {
			c.mu.Unlock()
			e.lastError = err
			return renderView(display, e, at)
		}
		if pending := c.inflight[location]; pending != nil {
			c.mu.Unlock()
			select {
			case <-pending.done:
				return renderView(display, pending.result, pending.completed)
			case <-ctx.Done():
				e.lastError = ctx.Err()
				return renderView(display, e, at)
			}
		}
		if at.Before(e.nextAttempt) {
			e.lastUsed = at
			c.entries[location] = e
			c.mu.Unlock()
			return renderView(display, e, at)
		}
		if _, exists := c.entries[location]; !exists && !c.makeRoom() {
			changed := c.changed
			c.mu.Unlock()
			select {
			case <-changed:
				continue
			case <-ctx.Done():
				e.lastError = ctx.Err()
				return renderView(display, e, at)
			}
		}
		if c.changed == nil {
			c.changed = make(chan struct{})
		}
		e.lastUsed = at
		c.entries[location] = e
		if c.inflight == nil {
			c.inflight = make(map[string]*refresh)
		}
		pending := &refresh{done: make(chan struct{})}
		c.inflight[location] = pending
		c.mu.Unlock()
		observation, err := fetch(ctx)
		c.mu.Lock()
		at = c.now()
		e.lastError = err
		if err != nil {
			e.nextAttempt = at.Add(retryInterval)
		} else {
			e.observation, e.available = observation, true
			e.nextAttempt = at.Add(refreshInterval)
		}
		e.lastUsed = at
		c.entries[location] = e
		delete(c.inflight, location)
		pending.result, pending.completed = e, at
		close(pending.done)
		close(c.changed)
		c.changed = make(chan struct{})
		c.mu.Unlock()
		return renderView(display, e, at)
	}
}

func renderView(location string, e entry, at time.Time) View {
	v := View{Location: location, Observation: e.observation, Available: e.available,
		Stale: !e.available || e.lastError != nil, NextAttempt: e.nextAttempt, LastError: e.lastError}
	if e.available {
		stale, err := e.observation.IsStale(at, maxObservationAge)
		v.Stale = v.Stale || stale || err != nil
		v.ConditionCode, v.ConditionText = weather.GetWeatherIconAt(e.observation.Code, at)
		// Generate for each render, rather than caching final SVG instance IDs.
		v.WeatherIcon = weather.GetSVGCodeByName(v.ConditionCode)
	}
	return v
}

// Called with mu held. Refreshing entries remain pinned until completion.
func (c *Cache) evictIdle(at time.Time) {
	for key, e := range c.entries {
		if c.inflight[key] == nil && at.Sub(e.lastUsed) >= cacheIdleTTL {
			delete(c.entries, key)
		}
	}
}

func (c *Cache) makeRoom() bool {
	if len(c.entries) < maxCacheEntries {
		return true
	}
	oldestKey := ""
	var oldest time.Time
	found := false
	for key, e := range c.entries {
		if c.inflight[key] != nil {
			continue
		}
		if !found || e.lastUsed.Before(oldest) {
			oldestKey, oldest, found = key, e.lastUsed, true
		}
	}
	if found {
		delete(c.entries, oldestKey)
	}
	return found
}
