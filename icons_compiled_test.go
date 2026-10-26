package weather

import (
	"strings"
	"testing"
)

func TestCompiledIconsMatchComposition(t *testing.T) {
	for key := range compiledIcons {
		t.Run(key, func(t *testing.T) {
			t.Parallel()
			data, err := iconFiles.ReadFile("icons/" + key + ".svg")
			if err != nil {
				t.Fatal(err)
			}
			for _, static := range []bool{false, true} {
				for _, prefix := range []string{"short", "fcw-template-marker", strings.Repeat("a", 64)} {
					want, ok := composeIconSVG(string(data), prefix, static)
					got, valid := getIconSVG(key, prefix, static)
					if !ok || !valid || got != want {
						t.Fatalf("compiled output differs: static=%t prefix=%s", static, prefix)
					}
				}
			}
		})
	}
}

func BenchmarkIconTemplateCold(b *testing.B) {
	data, _ := iconFiles.ReadFile("icons/clear-day.svg")
	b.ReportAllocs()
	for b.Loop() {
		benchmarkSVG, _ = composeIconSVG(string(data), "cold", false)
	}
}

func TestCompiledTemplateFootprint(t *testing.T) {
	bytes, fragments, sources := 0, 0, 0
	for key, icon := range compiledIcons {
		data, _ := iconFiles.ReadFile("icons/" + key + ".svg")
		sources += len(data)
		for _, compile := range []func() ([]string, bool){icon.animated, icon.static} {
			parts, ok := compile()
			if !ok {
				t.Fatal("template failed")
			}
			fragments += len(parts)
			for _, part := range parts {
				bytes += len(part)
			}
		}
	}
	t.Logf("all animated/static templates: %d fragment text bytes, %d fragments; retained template/source string payload approximately %d bytes, excluding slice headers, maps and closures", bytes, fragments, bytes+(fragments-2*len(compiledIcons))*len("fcw-template-marker")+sources)
}

func TestConcurrentTemplateInitialization(t *testing.T) {
	// Use fresh lazy templates so previous tests cannot hide initialization races.
	templates := newCompiledIcons()
	results := make(chan bool, 64)
	for i := 0; i < 64; i++ {
		go func(i int) {
			icon := templates["clear-day"]
			compile := icon.animated
			static := i%2 == 0
			if static {
				compile = icon.static
			}
			parts, ok := compile()
			data, _ := iconFiles.ReadFile("icons/clear-day.svg")
			prefix := strings.Repeat("x", i+1)
			want, valid := composeIconSVG(string(data), prefix, static)
			results <- ok && valid && strings.Join(parts, prefix) == want
		}(i)
	}
	for i := 0; i < 64; i++ {
		if !<-results {
			t.Fatal("concurrent template output mismatch")
		}
	}
}
