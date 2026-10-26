package weather

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestSearchLocationsPreservesProviderFields(t *testing.T) {
	client := fixtureClient(`{"status":200,"data":{"101250608":"湖南, 常德, 津市","101081213":"内蒙古, 阿拉善盟","101060306":"吉林, 延边","101010100":"北京, 北京"}}`)
	locations, err := searchLocations(t.Context(), "任意关键词", client)
	if err != nil {
		t.Fatal(err)
	}
	want := []Location{{"101010100", "北京", "北京", ""}, {"101060306", "吉林", "延边", ""}, {"101081213", "内蒙古", "阿拉善盟", ""}, {"101250608", "湖南", "常德", "津市"}}
	if len(locations) != len(want) {
		t.Fatalf("locations=%+v", locations)
	}
	for index, location := range locations {
		if location != want[index] {
			t.Errorf("location=%+v, want %+v", location, want[index])
		}
		roundtrip, err := parseProviderLocation(location.String())
		roundtrip.ID = location.ID
		if err != nil || roundtrip != location {
			t.Fatalf("roundtrip=%+v, %v", roundtrip, err)
		}
	}
}

func TestSearchLocationQueryEncoding(t *testing.T) {
	keyword := " 朝阳 &city=北京?县=1 "
	client := &http.Client{Transport: fixtureTransport(func(r *http.Request) (*http.Response, error) {
		if r.URL.Host != "i.news.qq.com" || r.URL.Path != "/city/like" {
			t.Fatalf("URL=%s", r.URL)
		}
		if q := r.URL.Query(); q.Get("city") != strings.TrimSpace(keyword) || q.Get("source") != "pc" || len(q) != 2 {
			t.Fatalf("query=%v", q)
		}
		return fixtureClient(`{"status":200,"data":{}}`).Transport.RoundTrip(r)
	})}
	got, err := searchLocations(t.Context(), keyword, client)
	if err != nil || len(got) != 0 {
		t.Fatalf("locations=%v err=%v", got, err)
	}
}

func TestSearchLocationFailures(t *testing.T) {
	for _, tc := range []struct{ name, body, stage string }{
		{"business", `{"status":500}`, "business"},
		{"decode", `{`, "decode"},
		{"missing city", `{"status":200,"data":{"1":"北京"}}`, "field"},
		{"empty province", `{"status":200,"data":{"1":",北京"}}`, "field"},
		{"empty county", `{"status":200,"data":{"1":"北京,北京,"}}`, "field"},
		{"too many levels", `{"status":200,"data":{"1":"a,b,c,d"}}`, "field"},
		{"empty code", `{"status":200,"data":{"":"北京,北京"}}`, "field"},
		{"wrong data type", `{"status":200,"data":[]}`, "decode"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := searchLocations(t.Context(), "北京", fixtureClient(tc.body))
			var detail *Error
			if !errors.As(err, &detail) || detail.Provider != "tencent" || detail.Stage != tc.stage {
				t.Fatalf("error=%v", err)
			}
		})
	}
}

func TestResolveLocationRequiresUniqueCandidate(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		cause      error
	}{
		{"none", `{"status":200,"data":{}}`, ErrLocationNotFound},
		{"null", `{"status":200,"data":null}`, ErrLocationNotFound},
		{"ambiguous", `{"status":200,"data":{"1":"北京,北京,朝阳","2":"辽宁,朝阳"}}`, ErrAmbiguousLocation},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			client := &http.Client{Transport: fixtureTransport(func(r *http.Request) (*http.Response, error) {
				calls++
				if r.URL.Path != "/city/like" {
					t.Fatalf("search failure queried weather: %s", r.URL)
				}
				return fixtureClient(tc.body).Transport.RoundTrip(r)
			})}
			got, err := getObservation(t.Context(), "朝阳", client)
			if !errors.Is(err, tc.cause) || got != (Observation{}) || calls != 1 {
				t.Fatalf("got=%+v err=%v calls=%d", got, err, calls)
			}
		})
	}
}

func TestWeatherUsesSearchResultCounty(t *testing.T) {
	calls := 0
	client := &http.Client{Transport: fixtureTransport(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.URL.Path == "/city/like" {
			if r.URL.Query().Get("city") != "津市" {
				t.Fatal(r.URL)
			}
			return fixtureClient(`{"status":200,"data":{"101250608":"湖南, 常德, 津市"}}`).Transport.RoundTrip(r)
		}
		q := r.URL.Query()
		if q.Get("province") != "湖南" || q.Get("city") != "常德" || q.Get("county") != "津市" || q.Get("weather_type") != "observe" {
			t.Fatalf("query=%v", q)
		}
		return fixtureClient(weatherFixture(nil)).Transport.RoundTrip(r)
	})}
	if _, err := getObservation(t.Context(), "津市", client); err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatalf("calls=%d, want search plus weather", calls)
	}
}

func TestSelectedLocationSkipsSearch(t *testing.T) {
	original := defaultHTTPClient
	t.Cleanup(func() { defaultHTTPClient = original })
	defaultHTTPClient = &http.Client{Transport: fixtureTransport(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/weather/common" || r.URL.Query().Get("county") != "朝阳" {
			t.Fatalf("URL=%s", r.URL)
		}
		return fixtureClient(weatherFixture(nil)).Transport.RoundTrip(r)
	})}
	location := Location{ID: "101010300", Province: "北京", City: "北京", County: "朝阳"}
	first, err := GetObservationAt(t.Context(), location)
	if err != nil {
		t.Fatal(err)
	}
	second, err := GetObservation(t.Context(), location.String())
	if err != nil || first != second {
		t.Fatalf("selected=%+v serialized=%+v err=%v", first, second, err)
	}
}

func TestSearchPublicAPIAndCancellation(t *testing.T) {
	original := defaultHTTPClient
	t.Cleanup(func() { defaultHTTPClient = original })
	defaultHTTPClient = fixtureClient(`{"status":200,"data":{"101010100":"北京, 北京"}}`)
	if got, err := SearchLocations(t.Context(), "北京"); err != nil || len(got) != 1 {
		t.Fatalf("got=%v err=%v", got, err)
	}
	for _, ctx := range []context.Context{nil, canceledLocationContext()} {
		if _, err := SearchLocations(ctx, "北京"); err == nil {
			t.Fatal("invalid context accepted")
		}
		if _, err := GetObservationAt(ctx, Location{Province: "北京", City: "北京"}); err == nil {
			t.Fatal("invalid context accepted")
		}
	}
}

func canceledLocationContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

func TestLocationValidationWithoutGeographicRules(t *testing.T) {
	for _, keyword := range []string{"", " ", "a\xff", "北\u200b京", "北\x00京"} {
		if _, err := searchLocations(t.Context(), keyword, fixtureClient("")); err == nil {
			t.Errorf("accepted %q", keyword)
		}
	}
	for _, location := range []Location{
		{}, {Province: "北京"}, {Province: "北京", City: "北\x00京"}, {Province: "北京", City: "北京", County: "a,b"}, {Province: "a,b", City: "北京"}, {Province: "北京", City: "北京", County: "\u200b"},
	} {
		if _, err := getObservationAt(t.Context(), location, fixtureClient("")); err == nil {
			t.Errorf("accepted %+v", location)
		}
	}
}

func TestIPLocationUsesTencentDirectory(t *testing.T) {
	for _, tc := range []struct {
		name, directory string
		cause           error
		want            string
	}{
		{"county", `{"status":200,"data":{"101250608":"湖南, 常德, 津市"}}`, nil, "湖南, 常德, 津市"},
		{"other province", `{"status":200,"data":{"1":"湖北, 城市"}}`, ErrLocationNotFound, ""},
		{"ambiguous", `{"status":200,"data":{"1":"湖南,甲,津市","2":"湖南,乙,津市"}}`, ErrAmbiguousLocation, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := &http.Client{Transport: fixtureTransport(func(r *http.Request) (*http.Response, error) {
				body := `{"ret":"ok","data":{"location":["中国","湖南","津市"]}}`
				if r.URL.Path == "/city/like" {
					body = tc.directory
				}
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
			})}
			got, err := getMyIPLocation(t.Context(), client)
			if !errors.Is(err, tc.cause) || got != tc.want {
				t.Fatalf("got=%q err=%v", got, err)
			}
		})
	}
}

func FuzzProviderLocation(f *testing.F) {
	for _, value := range []string{"北京,北京", "湖南,常德,津市", "", "a,b,c,d", "a,\u200b"} {
		f.Add(value)
	}
	f.Fuzz(func(t *testing.T, value string) {
		got, err := parseProviderLocation(value)
		if err != nil {
			return
		}
		again, err := parseProviderLocation(got.String())
		if err != nil || got != again {
			t.Fatalf("roundtrip=%+v err=%v", again, err)
		}
	})
}

func TestNoLocalAddressRewrite(t *testing.T) {
	keyword := "广西壮族自治区南宁市"
	client := &http.Client{Transport: fixtureTransport(func(r *http.Request) (*http.Response, error) {
		if r.URL.Query().Get("city") != keyword {
			t.Fatalf("keyword rewritten: %s", r.URL)
		}
		return fixtureClient(`{"status":200,"data":{}}`).Transport.RoundTrip(r)
	})}
	if _, err := getObservation(t.Context(), keyword, client); !errors.Is(err, ErrLocationNotFound) {
		t.Fatalf("error=%v", err)
	}
}

func TestSearchFailurePreservesCause(t *testing.T) {
	cause := errors.New("directory unavailable")
	client := &http.Client{Transport: fixtureTransport(func(*http.Request) (*http.Response, error) { return nil, cause })}
	if _, err := searchLocations(t.Context(), "北京", client); !errors.Is(err, cause) {
		t.Fatalf("error=%v", err)
	}
}

func TestCancellationBetweenSearchAndWeather(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	calls := 0
	client := &http.Client{Transport: fixtureTransport(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.Context() != ctx {
			t.Fatal("request lost context")
		}
		cancel()
		return fixtureClient(`{"status":200,"data":{"101010100":"北京,北京"}}`).Transport.RoundTrip(r)
	})}
	if _, err := getObservation(ctx, "北京", client); !errors.Is(err, context.Canceled) {
		t.Fatalf("error=%v", err)
	}
	if calls != 1 {
		t.Fatalf("canceled request queried weather; calls=%d", calls)
	}
}

func TestIPDirectoryFailurePropagates(t *testing.T) {
	client := &http.Client{Transport: fixtureTransport(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/city/like" {
			return fixtureClient(`{"status":503}`).Transport.RoundTrip(r)
		}
		return fixtureClient(`{"ret":"ok","data":{"location":["中国","湖南","津市"]}}`).Transport.RoundTrip(r)
	})}
	got, err := getMyIPLocation(t.Context(), client)
	var detail *Error
	if got != "" || !errors.As(err, &detail) || detail.Provider != "tencent" || detail.Stage != "business" {
		t.Fatalf("got=%q err=%v", got, err)
	}
}
