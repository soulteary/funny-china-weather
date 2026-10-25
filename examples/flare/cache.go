// Package flare demonstrates a Flare-style weather consumer with bounded retries.
// It is example application code; the weather library itself has no cache policy.
package flare

import (
	"context"
	"html/template"
	"sync"
	"time"

	weather "github.com/soulteary/funny-china-weather"
)

const (
	refreshInterval   = 10 * time.Minute
	retryInterval     = time.Minute
	maxObservationAge = 30 * time.Minute
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
}

// Cache keeps separate data for each location. A mutex also prevents concurrent
// page requests from starting duplicate refreshes. Requests can wait for an
// in-progress refresh (the provider adapter has the library's three-second cap).
type Cache struct {
	mu      sync.Mutex
	entries map[string]entry
	fetch   func(context.Context, string) (weather.Observation, error)
	now     func() time.Time
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
	c.mu.Lock()
	defer c.mu.Unlock()
	at := c.now()
	e := c.entries[location]
	if !at.Before(e.nextAttempt) {
		observation, err := c.fetch(ctx, location)
		at = c.now() // Schedule from completion, including network delay.
		e.lastError = err
		if err != nil {
			e.nextAttempt = at.Add(retryInterval)
		} else {
			e.observation, e.available = observation, true
			e.nextAttempt = at.Add(refreshInterval)
		}
		c.entries[location] = e
	}
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
