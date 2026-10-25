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

// GetMyIPLocation returns the province and city associated with the outgoing IP.
func GetMyIPLocation() (string, error) {
	return GetMyIPLocationContext(context.Background())
}

// GetMyIPLocationContext returns the province and city associated with the outgoing
// IP while respecting ctx and the shared three-second client timeout. ctx must not
// be nil. An error returns an empty location.
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

	parsed, err := normalizeLocationFields(location[1], location[2])
	if err != nil {
		return "", &Error{Provider: "ipip", Stage: "field", Field: "data.location", Err: fmt.Errorf("invalid data.location: %w", err)}
	}
	return parsed.String(), nil
}
