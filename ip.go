package weather

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type responseIPIP struct {
	Ret  string `json:"ret"`
	Data struct {
		IP       string   `json:"ip"`
		Location []string `json:"location"`
	} `json:"data"`
}

// GetMyIPLocation resolves the outgoing IP city through Tencent and returns a
// comma-separated provider address accepted by the string weather APIs.
func GetMyIPLocation() (string, error) {
	return GetMyIPLocationContext(context.Background())
}

// GetMyIPLocationContext looks up the IPIP city in Tencent's directory and
// returns the selected Location.String() address. Province names must match
// exactly across providers; no match or multiple matches return an error.
// Both requests respect ctx and the shared three-second per-request timeout.
// ctx must not be nil. An error returns an empty location.
func GetMyIPLocationContext(ctx context.Context) (string, error) {
	return getMyIPLocation(ctx, defaultHTTPClient)
}

func getMyIPLocation(ctx context.Context, client httpDoer) (string, error) {
	response := new(responseIPIP)
	if err := getJSONWithClient(ctx, client, "ipip", "https://myip.ipip.net/json", response); err != nil {
		return "", err
	}

	if response.Ret != "ok" {
		return "", &Error{Provider: "ipip", Stage: "business", BusinessCode: response.Ret, Err: fmt.Errorf("unexpected ret %q", response.Ret)}
	}

	location := response.Data.Location
	if len(location) < 3 {
		return "", &Error{Provider: "ipip", Stage: "field", Field: "data.location", Err: fmt.Errorf("data.location must contain country, province, and city; got %d entries", len(location))}
	}
	for index, field := range []string{"country", "province", "city"} {
		location[index] = strings.TrimSpace(location[index])
		if location[index] == "" {
			return "", &Error{Provider: "ipip", Stage: "field", Field: "data.location." + field, Err: fmt.Errorf("data.location %s is empty", field)}
		}
	}
	if location[0] != "中国" {
		return "", &Error{Provider: "ipip", Stage: "field", Field: "data.location.country", Err: errors.New("非境内IP地址，放弃继续解析 :(")}
	}

	// Search the city verbatim and use the province only to disambiguate exact
	// provider fields. Never invent a suffix or choose a different province.
	locations, err := searchLocations(ctx, location[2], client)
	if err != nil {
		return "", err
	}
	var matches []Location
	for _, candidate := range locations {
		if candidate.Province == location[1] {
			matches = append(matches, candidate)
		}
	}
	selected, err := uniqueLocation(matches)
	if err != nil {
		return "", err
	}
	return selected.String(), nil
}
