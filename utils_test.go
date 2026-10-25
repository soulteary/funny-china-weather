package weather

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type trackedResponseBody struct {
	io.Reader
	closed bool
}

func (body *trackedResponseBody) Close() error {
	body.closed = true
	return nil
}

type failedReader struct{ err error }

func (reader failedReader) Read([]byte) (int, error) { return 0, reader.err }

func TestGetJSONResponses(t *testing.T) {
	valid := `{"value":"ok"}`
	tests := []struct {
		name      string
		status    int
		body      string
		wantError string
		wantCause error
	}{
		{name: "success", status: 200, body: valid},
		{name: "other 2xx", status: 201, body: valid},
		{name: "unknown fields", status: 200, body: `{"value":"ok","future":{"field":true}}`},
		{name: "trailing whitespace", status: 200, body: valid + " \n\t"},
		{name: "exact body limit", status: 200, body: valid + strings.Repeat(" ", maxJSONResponseBytes-len(valid))},
		{name: "body too large", status: 200, body: valid + strings.Repeat(" ", maxJSONResponseBytes-len(valid)+1), wantError: "response exceeds", wantCause: errResponseTooLarge},
		{name: "size checked before JSON", status: 200, body: strings.Repeat("x", maxJSONResponseBytes+1), wantError: "response exceeds", wantCause: errResponseTooLarge},
		{name: "client HTTP error", status: 400, body: valid, wantError: "HTTP status 400"},
		{name: "HTTP before JSON", status: 503, body: `{`, wantError: "HTTP status 503"},
		{name: "redirect without location", status: 302, body: valid, wantError: "HTTP status 302"},
		{name: "empty response", status: 200, body: "", wantError: "decode response"},
		{name: "malformed JSON", status: 200, body: `{`, wantError: "decode response"},
		{name: "multiple JSON values", status: 200, body: valid + valid, wantError: "decode response"},
		{name: "garbage suffix", status: 200, body: valid + "x", wantError: "decode response"},
		{name: "wrong field type", status: 200, body: `{"value":1}`, wantError: "decode response"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body := &trackedResponseBody{Reader: strings.NewReader(test.body)}
			type contextKey struct{}
			ctx := context.WithValue(context.Background(), contextKey{}, "request context")
			client := &http.Client{Transport: fixtureTransport(func(request *http.Request) (*http.Response, error) {
				if request.Method != http.MethodGet || request.Context().Value(contextKey{}) != "request context" {
					t.Fatalf("request method/context = %s/%v", request.Method, request.Context().Value(contextKey{}))
				}
				return &http.Response{StatusCode: test.status, Header: make(http.Header), Body: body, Request: request}, nil
			})}
			var target struct {
				Value string `json:"value"`
			}
			err := getJSONWithClient(ctx, client, "fixture", "https://weather.example.test/data", &target)
			if !body.closed {
				t.Error("response body was not closed")
			}
			if test.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), "fixture:") || !strings.Contains(err.Error(), test.wantError) {
					t.Fatalf("error = %v, want provider and %q", err, test.wantError)
				}
				if test.wantCause != nil && !errors.Is(err, test.wantCause) {
					t.Fatalf("error = %v, want cause %v", err, test.wantCause)
				}
				return
			}
			if err != nil || target.Value != "ok" {
				t.Fatalf("response = %q, error = %v", target.Value, err)
			}
		})
	}
}

func TestGetJSONReadFailure(t *testing.T) {
	readError := errors.New("body read failed")
	body := &trackedResponseBody{Reader: failedReader{err: readError}}
	client := &http.Client{Transport: fixtureTransport(func(request *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: body, Request: request}, nil
	})}
	var target interface{}
	err := getJSONWithClient(context.Background(), client, "fixture", "https://weather.example.test/data", &target)
	if !errors.Is(err, readError) || !strings.Contains(err.Error(), "fixture: read response") || !body.closed {
		t.Fatalf("error = %v, body closed = %v; want preserved read error and closed body", err, body.closed)
	}
}

func TestGetJSONInvalidRequest(t *testing.T) {
	client := &http.Client{Transport: fixtureTransport(func(*http.Request) (*http.Response, error) {
		t.Fatal("invalid URL reached transport")
		return nil, nil
	})}
	var target interface{}
	err := getJSONWithClient(context.Background(), client, "fixture", "://invalid", &target)
	if err == nil || !strings.Contains(err.Error(), "fixture: create request") {
		t.Fatalf("error = %v, want request creation error", err)
	}
}

func TestProviderContextErrors(t *testing.T) {
	providers := []struct {
		name string
		call func(context.Context, httpDoer) error
	}{
		{name: "ipip", call: func(ctx context.Context, client httpDoer) error { _, err := getMyIPLocation(ctx, client); return err }},
		{name: "tencent", call: func(ctx context.Context, client httpDoer) error {
			_, _, _, _, err := getWeatherByLocation(ctx, "北京市", client)
			return err
		}},
	}
	for _, provider := range providers {
		for _, cause := range []error{context.Canceled, context.DeadlineExceeded} {
			t.Run(provider.name+"/"+cause.Error(), func(t *testing.T) {
				var ctx context.Context
				var cancel context.CancelFunc
				if cause == context.Canceled {
					ctx, cancel = context.WithCancel(context.Background())
					cancel()
				} else {
					ctx, cancel = context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
					defer cancel()
				}
				client := &http.Client{Transport: fixtureTransport(func(request *http.Request) (*http.Response, error) { return nil, request.Context().Err() })}
				err := provider.call(ctx, client)
				if !errors.Is(err, cause) || !strings.Contains(err.Error(), provider.name+": request failed") {
					t.Fatalf("error = %v, want provider and cause %v", err, cause)
				}
			})
		}
	}
}

func TestGetJSONPreservesDecodeCause(t *testing.T) {
	var target struct {
		Value string `json:"value"`
	}
	err := getJSONWithClient(context.Background(), fixtureClient(`{"value":1}`), "fixture", "https://weather.example.test/data", &target)
	var cause *json.UnmarshalTypeError
	if !errors.As(err, &cause) || cause.Field != "value" {
		t.Fatalf("error = %v, want retained JSON field type error", err)
	}
}
