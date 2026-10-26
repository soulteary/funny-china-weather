//go:build integration

package weather

import (
	"context"
	"testing"
	"time"
)

func TestTencentLocationIntegration(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()
	for _, keyword := range []string{"北京", "南宁", "津市"} {
		t.Run(keyword, func(t *testing.T) {
			locations, err := SearchLocations(ctx, keyword)
			if err != nil || len(locations) != 1 {
				t.Fatalf("locations=%+v err=%v", locations, err)
			}
			selected := locations[0]
			if keyword == "津市" && (selected.City != "常德" || selected.County != "津市") {
				t.Fatalf("unexpected county hierarchy: %+v", selected)
			}
			observation, err := GetObservationAt(ctx, selected)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := ParseUpdateTime(observation.UpdateTime); err != nil {
				t.Fatal(err)
			}
			t.Logf("%s -> %s (%s); code=%d", keyword, selected.String(), selected.ID, observation.Code)
		})
	}
}
