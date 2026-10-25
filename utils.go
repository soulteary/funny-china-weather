package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const maxJSONResponseBytes = 1 << 20

var defaultHTTPClient = &http.Client{Timeout: 3 * time.Second}

type httpDoer interface {
	Do(*http.Request) (*http.Response, error)
}

func getJSONWithClient(ctx context.Context, client httpDoer, provider, requestURL string, target interface{}) error {
	if ctx == nil {
		return &Error{Provider: provider, Stage: "input", Field: "context", Err: fmt.Errorf("context must not be nil")}
	}
	if err := ctx.Err(); err != nil {
		return &Error{Provider: provider, Stage: "request", Err: fmt.Errorf("request failed: %w", err)}
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return &Error{Provider: provider, Stage: "request", Err: fmt.Errorf("create request: %w", err)}
	}
	response, err := client.Do(request)
	if err != nil {
		return &Error{Provider: provider, Stage: "request", Err: fmt.Errorf("request failed: %w", err)}
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return &Error{Provider: provider, Stage: "http", HTTPStatus: response.StatusCode}
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, maxJSONResponseBytes+1))
	if err != nil {
		return &Error{Provider: provider, Stage: "read", Err: fmt.Errorf("read response: %w", err)}
	}
	if len(body) > maxJSONResponseBytes {
		return &Error{Provider: provider, Stage: "size", Err: fmt.Errorf("response exceeds %d bytes: %w", maxJSONResponseBytes, ErrResponseTooLarge)}
	}
	if err := json.Unmarshal(body, target); err != nil {
		return &Error{Provider: provider, Stage: "decode", Err: fmt.Errorf("decode response: %w", err)}
	}
	return nil
}
