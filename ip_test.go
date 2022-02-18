package weather

import (
	"testing"
)

func TestGetMyIPLocation(t *testing.T) {
	t.Run("Test `GetMyIPLocation` API is alive", func(t *testing.T) {
		_, err := GetMyIPLocation()
		if err != nil {
			t.Errorf("`myip.ipip.net` error '%s'", err)
		}
	})
}
