package weather

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestGetMyIPLocation(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		want      string
		wantError bool
	}{
		{name: "province and city", body: `{"ret":"ok","data":{"location":["中国","浙江","杭州"]}}`, want: "浙江, 杭州"},
		{name: "Beijing municipality", body: `{"ret":"ok","data":{"location":["中国","北京","北京"]}}`, want: "北京, 北京"},
		{name: "Shanghai municipality", body: `{"ret":"ok","data":{"location":["中国","上海","上海"]}}`, want: "上海, 上海"},
		{name: "Chongqing municipality", body: `{"ret":"ok","data":{"location":["中国","重庆","重庆"]}}`, want: "重庆, 重庆"},
		{name: "Tianjin municipality", body: `{"ret":"ok","data":{"location":["中国","天津","天津"]}}`, want: "天津, 天津"},
		{name: "surrounding whitespace", body: `{"ret":"ok","data":{"location":[" 中国 "," 浙江 "," 杭州 "]}}`, want: "浙江, 杭州"},
		{name: "missing data", body: `{"ret":"ok"}`, wantError: true},
		{name: "missing location", body: `{"ret":"ok","data":{}}`, wantError: true},
		{name: "null location", body: `{"ret":"ok","data":{"location":null}}`, wantError: true},
		{name: "empty location", body: `{"ret":"ok","data":{"location":[]}}`, wantError: true},
		{name: "country only", body: `{"ret":"ok","data":{"location":["中国"]}}`, wantError: true},
		{name: "missing city", body: `{"ret":"ok","data":{"location":["中国","浙江"]}}`, wantError: true},
		{name: "empty country", body: `{"ret":"ok","data":{"location":["","浙江","杭州"]}}`, wantError: true},
		{name: "empty province", body: `{"ret":"ok","data":{"location":["中国"," ","杭州"]}}`, wantError: true},
		{name: "empty city", body: `{"ret":"ok","data":{"location":["中国","浙江",""]}}`, wantError: true},
		{name: "foreign location", body: `{"ret":"ok","data":{"location":["日本","东京都","东京"]}}`, wantError: true},
		{name: "malformed JSON", body: `{"data":`, wantError: true},
		{name: "wrong location type", body: `{"ret":"ok","data":{"location":"中国"}}`, wantError: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := getMyIPLocation(context.Background(), ipFixtureClient(test.body))
			if (err != nil) != test.wantError {
				t.Fatalf("getMyIPLocation() error = %v, wantError %v", err, test.wantError)
			}
			if got != test.want {
				t.Errorf("getMyIPLocation() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestGetMyIPLocationRequestFailure(t *testing.T) {
	requestError := errors.New("request failed")
	client := &http.Client{Transport: fixtureTransport(func(*http.Request) (*http.Response, error) {
		return nil, requestError
	})}
	location, err := getMyIPLocation(context.Background(), client)
	if !errors.Is(err, requestError) || location != "" {
		t.Fatalf("getMyIPLocation() = (%q, %v), want empty location and request error", location, err)
	}
}

func TestGetMyIPLocationBusinessStatus(t *testing.T) {
	tests := []struct{ name, body string }{
		{name: "failure with valid location", body: `{"ret":"error","data":{"location":["中国","浙江","杭州"]}}`},
		{name: "missing ret", body: `{"data":{"location":["中国","浙江","杭州"]}}`},
		{name: "null ret", body: `{"ret":null,"data":{"location":["中国","浙江","杭州"]}}`},
		{name: "business before data", body: `{"ret":"error"}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			location, err := getMyIPLocation(context.Background(), ipFixtureClient(test.body))
			if err == nil || location != "" || !strings.Contains(err.Error(), "ipip: unexpected ret") {
				t.Fatalf("location/error = %q/%v, want empty location and ret error", location, err)
			}
		})
	}
}
