package weather

import (
	"io"
	"net/http"
	"strings"
)

type fixtureTransport func(*http.Request) (*http.Response, error)

func (transport fixtureTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	return transport(request)
}

func fixtureClient(body string) *http.Client {
	return &http.Client{Transport: fixtureTransport(func(request *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    request,
		}, nil
	})}
}

// IP response fixtures use a separate, fixed Tencent directory response.
func ipFixtureClient(body string) *http.Client {
	ip := fixtureClient(body)
	directory := fixtureClient(`{"status":200,"data":{"101010100":"北京, 北京","101210101":"浙江, 杭州","101020100":"上海, 上海","101030100":"天津, 天津","101040100":"重庆, 重庆"}}`)
	return &http.Client{Transport: fixtureTransport(func(request *http.Request) (*http.Response, error) {
		if request.URL.Path == "/city/like" {
			return directory.Transport.RoundTrip(request)
		}
		return ip.Transport.RoundTrip(request)
	})}
}
