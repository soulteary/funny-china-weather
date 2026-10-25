package weather

import (
	"errors"
	"fmt"
)

// ErrResponseTooLarge indicates that an upstream response exceeds the 1 MiB limit.
// Use errors.Is to identify it through a wrapping *Error.
var ErrResponseTooLarge = errors.New("response too large")

var errResponseTooLarge = ErrResponseTooLarge

// Error describes an input, request, or provider response failure.
// Use errors.As to inspect its fields and errors.Is to inspect the underlying cause.
// Stage is one of input, request, http, read, size, decode, business, field, or freshness.
// HTTPStatus is set for HTTP status failures; BusinessCode contains the provider's
// ret or status value for business failures. Other unavailable fields remain zero.
// Error messages are descriptive and should not be used for programmatic matching.
type Error struct {
	// Provider is "ipip" or "tencent" for errors from those providers.
	Provider string
	// Stage identifies the failed operation using the stable values listed above.
	Stage string
	// HTTPStatus is the rejected HTTP response status, or zero when unavailable.
	HTTPStatus int
	// BusinessCode is IPIP's ret or Tencent's decimal status, when rejected.
	BusinessCode string
	// Field identifies invalid input or response data, when applicable.
	Field string
	// Err is the wrapped cause and remains available through Unwrap.
	Err error
}

// Error returns a human-readable description of the failure.
func (err *Error) Error() string {
	if err == nil {
		return "<nil>"
	}
	message := err.Stage
	if message == "" {
		message = "weather error"
	}
	switch {
	case err.Err != nil:
		message = err.Err.Error()
	case err.HTTPStatus != 0:
		message = fmt.Sprintf("unexpected HTTP status %d", err.HTTPStatus)
	case err.BusinessCode != "":
		message = fmt.Sprintf("unexpected business code %q", err.BusinessCode)
	}
	if err.Provider == "" {
		return message
	}
	return err.Provider + ": " + message
}

// Unwrap returns the original cause, when one is available.
func (err *Error) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.Err
}
