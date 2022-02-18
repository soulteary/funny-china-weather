package weather

import (
	"encoding/json"
	"net/http"
	"time"
)

func getJSON(url string, timeout time.Duration, target interface{}) error {
	var client = &http.Client{Timeout: timeout * time.Second}
	r, err := client.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}
