package weather

import "testing"

var benchmarkSVG string

func BenchmarkIconSVG(b *testing.B) {
	for _, key := range []string{"clear-day", "cloudy-day", "rain-heavy", "unknown"} {
		b.Run(key, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				benchmarkSVG, _ = GetIconSVG(key)
			}
		})
	}
}
