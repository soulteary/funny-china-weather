package weather

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const maxJSONResponseBytes = 1 << 20

var (
	defaultHTTPClient   = &http.Client{Timeout: 3 * time.Second}
	errResponseTooLarge = errors.New("response too large")
)

type httpDoer interface {
	Do(*http.Request) (*http.Response, error)
}

func getJSONWithClient(ctx context.Context, client httpDoer, provider, requestURL string, target interface{}) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return fmt.Errorf("%s: create request: %w", provider, err)
	}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("%s: request failed: %w", provider, err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("%s: unexpected HTTP status %d", provider, response.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, maxJSONResponseBytes+1))
	if err != nil {
		return fmt.Errorf("%s: read response: %w", provider, err)
	}
	if len(body) > maxJSONResponseBytes {
		return fmt.Errorf("%s: response exceeds %d bytes: %w", provider, maxJSONResponseBytes, errResponseTooLarge)
	}
	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("%s: decode response: %w", provider, err)
	}
	return nil
}
