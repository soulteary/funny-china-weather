package weather

import (
	"encoding/json"
	"net/http"
	"time"
)

type httpDoer interface {
	Do(*http.Request) (*http.Response, error)
}

func newHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{Timeout: timeout * time.Second}
}

func getJSON(url string, timeout time.Duration, target interface{}) error {
	return getJSONWithClient(newHTTPClient(timeout), url, target)
}

func getJSONWithClient(client httpDoer, url string, target interface{}) error {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	return json.NewDecoder(response.Body).Decode(target)
}
