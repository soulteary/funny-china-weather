package weather

import (
	"errors"
	"testing"
)

func TestErrorFormatting(t *testing.T) {
	cause := errors.New("underlying failure")
	for _, tc := range []struct {
		name   string
		detail *Error
		want   string
	}{
		{"nil", nil, "<nil>"},
		{"empty", &Error{}, "weather error"},
		{"stage", &Error{Stage: "read"}, "read"},
		{"provider", &Error{Provider: "ipip"}, "ipip: weather error"},
		{"http", &Error{HTTPStatus: 503}, "unexpected HTTP status 503"},
		{"business", &Error{BusinessCode: "limited"}, `unexpected business code "limited"`},
		{"cause precedence", &Error{Provider: "tencent", HTTPStatus: 503, BusinessCode: "limited", Err: cause}, "tencent: underlying failure"},
		{"HTTP precedence", &Error{HTTPStatus: 503, BusinessCode: "limited"}, "unexpected HTTP status 503"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.detail.Error(); got != tc.want {
				t.Errorf("Error()=%q, want %q", got, tc.want)
			}
			var want error
			if tc.detail != nil {
				want = tc.detail.Err
			}
			if got := tc.detail.Unwrap(); got != want {
				t.Errorf("Unwrap()=%v, want %v", got, want)
			}
		})
	}
}
