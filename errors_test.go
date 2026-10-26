package weather

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

func TestStructuredErrors(t *testing.T) {
	httpFailure := func() error {
		client := &http.Client{Transport: fixtureTransport(func(request *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 429, Header: make(http.Header), Body: http.NoBody, Request: request}, nil
		})}
		_, err := getObservation(context.Background(), "北京, 北京", client)
		return err
	}
	tests := []struct {
		name, provider, stage, field, code string
		httpStatus                         int
		call                               func() error
	}{
		{name: "input", provider: "tencent", stage: "input", field: "location", call: func() error { _, err := GetObservation(context.Background(), ""); return err }},
		{name: "rate limit", provider: "tencent", stage: "http", httpStatus: 429, call: httpFailure},
		{name: "Tencent business", provider: "tencent", stage: "business", code: "429", call: func() error {
			_, err := getObservation(context.Background(), "北京, 北京", fixtureClient(`{"status":429,"message":"limited"}`))
			return err
		}},
		{name: "IPIP business", provider: "ipip", stage: "business", code: "error", call: func() error {
			_, err := getMyIPLocation(context.Background(), fixtureClient(`{"ret":"error"}`))
			return err
		}},
		{name: "missing IP data", provider: "ipip", stage: "field", field: "data.location", call: func() error {
			_, err := getMyIPLocation(context.Background(), fixtureClient(`{"ret":"ok"}`))
			return err
		}},
		{name: "IP country", provider: "ipip", stage: "field", field: "data.location.country", call: func() error {
			_, err := getMyIPLocation(context.Background(), fixtureClient(`{"ret":"ok","data":{"location":["日本","東京","東京"]}}`))
			return err
		}},
		{name: "numeric field", provider: "tencent", stage: "field", field: "degree", call: func() error {
			_, err := getObservation(context.Background(), "北京, 北京", fixtureClient(weatherFixture(map[string]string{"degree": "invalid"})))
			return err
		}},
		{name: "decode", provider: "tencent", stage: "decode", call: func() error {
			_, err := getObservation(context.Background(), "北京, 北京", fixtureClient(`{`))
			return err
		}},
		{name: "response limit", provider: "tencent", stage: "size", call: func() error {
			_, err := getObservation(context.Background(), "北京, 北京", fixtureClient(strings.Repeat("x", maxJSONResponseBytes+1)))
			return err
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.call()
			var detail *Error
			if !errors.As(err, &detail) {
				t.Fatalf("error = %v, want *Error", err)
			}
			if detail.Provider != test.provider || detail.Stage != test.stage || detail.Field != test.field || detail.HTTPStatus != test.httpStatus || detail.BusinessCode != test.code {
				t.Fatalf("error metadata = %#v", detail)
			}
			checkErrorCause(t, err, detail, test.stage, test.field)
		})
	}
}

func checkErrorCause(t *testing.T, err error, detail *Error, stage, field string) {
	t.Helper()
	switch stage {
	case "input":
		if detail.Err == nil || errors.Unwrap(detail) == nil {
			t.Fatal("input error lost underlying cause")
		}
	case "size":
		if !errors.Is(err, ErrResponseTooLarge) || !errors.Is(err, errResponseTooLarge) {
			t.Fatalf("size error lost sentinel: %v", err)
		}
	case "field":
		if field == "degree" {
			var cause *strconv.NumError
			if !errors.As(err, &cause) || !errors.Is(err, strconv.ErrSyntax) {
				t.Fatalf("numeric cause lost: %v", err)
			}
		}
	case "decode":
		var cause *json.SyntaxError
		if !errors.As(err, &cause) {
			t.Fatalf("JSON cause lost: %v", err)
		}
	}
}
