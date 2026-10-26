package weather

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ErrLocationNotFound means Tencent returned no matching locations.
var ErrLocationNotFound = errors.New("location not found")

// ErrAmbiguousLocation means a query matched multiple locations. Use
// SearchLocations and GetObservationAt to select a specific candidate.
var ErrAmbiguousLocation = errors.New("ambiguous location")

// Location contains Tencent's location names, without administrative-name
// normalization. County is optional; ID is the provider's weather city code.
// Obtain locations from SearchLocations rather than constructing names locally.
type Location struct {
	ID       string
	Province string
	City     string
	County   string
}

// String returns a comma-separated address accepted by the string weather APIs.
// It retains both province and city even when they are equal.
func (location Location) String() string {
	parts := []string{location.Province, location.City}
	if location.County != "" {
		parts = append(parts, location.County)
	}
	return strings.Join(parts, ", ")
}

// SearchLocations searches Tencent's location directory using the supplied
// keyword, for example 北京, 南宁, or 津市. Results are sorted by city code.
// No suffixes or province aliases are added. An empty result is not an error.
func SearchLocations(ctx context.Context, keyword string) ([]Location, error) {
	return searchLocations(ctx, keyword, defaultHTTPClient)
}

func searchLocations(ctx context.Context, keyword string, client httpDoer) ([]Location, error) {
	keyword = strings.TrimSpace(keyword)
	if !validLocationText(keyword) {
		return nil, locationError("input", "location", fmt.Errorf("location keyword must be nonempty valid text"))
	}
	query := url.Values{"source": {"pc"}, "city": {keyword}}
	var response struct {
		Status  int               `json:"status"`
		Message string            `json:"message"`
		Data    map[string]string `json:"data"`
	}
	if err := getJSONWithClient(ctx, client, "tencent", "https://i.news.qq.com/city/like?"+query.Encode(), &response); err != nil {
		return nil, err
	}
	if response.Status != 200 {
		return nil, &Error{Provider: "tencent", Stage: "business", BusinessCode: strconv.Itoa(response.Status), Err: fmt.Errorf("unexpected status %d: %s", response.Status, response.Message)}
	}
	ids := make([]string, 0, len(response.Data))
	for id := range response.Data {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	locations := make([]Location, 0, len(ids))
	for _, id := range ids {
		location, err := parseProviderLocation(response.Data[id])
		if err != nil {
			return nil, locationError("field", "data", fmt.Errorf("invalid location %q: %w", id, err))
		}
		if strings.TrimSpace(id) == "" {
			return nil, locationError("field", "data", fmt.Errorf("empty location code"))
		}
		location.ID = id
		locations = append(locations, location)
	}
	return locations, nil
}

func locationError(stage, field string, cause error) error {
	return &Error{Provider: "tencent", Stage: stage, Field: field, Err: cause}
}

// Only validate text and field boundaries; Tencent owns geographic semantics.
func validLocationText(value string) bool {
	if strings.TrimSpace(value) == "" || !utf8.ValidString(value) {
		return false
	}
	for _, char := range value {
		if unicode.IsControl(char) || unicode.Is(unicode.Cf, char) {
			return false
		}
	}
	return true
}

func parseProviderLocation(value string) (Location, error) {
	parts := strings.Split(value, ",")
	if len(parts) < 2 || len(parts) > 3 {
		return Location{}, fmt.Errorf("expected province, city, and optional county")
	}
	for index := range parts {
		parts[index] = strings.TrimSpace(parts[index])
		if !validLocationText(parts[index]) {
			return Location{}, fmt.Errorf("invalid location field")
		}
	}
	location := Location{Province: parts[0], City: parts[1]}
	if len(parts) == 3 {
		location.County = parts[2]
	}
	return location, nil
}

func resolveLocation(ctx context.Context, value string, client httpDoer) (Location, error) {
	// This is serialization of provider fields, not Chinese address parsing.
	if strings.Contains(value, ",") {
		location, err := parseProviderLocation(value)
		if err != nil {
			return Location{}, locationError("input", "location", err)
		}
		return location, nil
	}
	locations, err := searchLocations(ctx, value, client)
	if err != nil {
		return Location{}, err
	}
	return uniqueLocation(locations)
}

func uniqueLocation(locations []Location) (Location, error) {
	switch len(locations) {
	case 0:
		return Location{}, locationError("field", "location", ErrLocationNotFound)
	case 1:
		return locations[0], nil
	default:
		return Location{}, locationError("field", "location", fmt.Errorf("%d candidates; use SearchLocations to select one: %w", len(locations), ErrAmbiguousLocation))
	}
}
