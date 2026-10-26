package weather

import (
	"strings"
	"testing"
)

func TestSVGCSSLiterals(t *testing.T) {
	names := svgNames{ids: map[string]string{"abc": "scope-abc"}, css: map[string]string{"pulse": "scope-pulse"}}
	for _, tc := range []struct{ name, input, want string }{
		{"quoted", `.pulse{content:"pulse";animation:pulse}`, `.scope-pulse{content:"pulse";animation:scope-pulse}`},
		{"escaped quote", `.pulse{content:'it\'s pulse';animation:pulse}`, `.scope-pulse{content:'it\'s pulse';animation:scope-pulse}`},
		{"trailing escape", `"pulse\`, `"pulse\`},
		{"unterminated quote", `'pulse`, `'pulse`},
		{"unterminated comment", `/* pulse`, `/* pulse`},
		{"identifier boundaries", `.pulse-long{animation:pulse}`, `.pulse-long{animation:scope-pulse}`},
		{"unknown URL", `.pulse{fill:url(#missing)}`, `.scope-pulse{fill:url(#missing)}`},
		{"selector and color", `#abc{fill:#abc}`, `#scope-abc{fill:#abc}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := rewriteSVGCSS(tc.input, names, false); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestSVGAccessibilityAndRootID(t *testing.T) {
	source := `<svg id="root" aria-labelledby="title external" aria-describedby="desc"><title id="title">A &amp; B</title><desc id="desc">Weather</desc><g class="one two"/><use href="https://example.test/#title"/></svg>`
	got, ok := composeIconSVG(source, "scope", false)
	if !ok {
		t.Fatal("composition failed")
	}
	for _, want := range []string{`id="scope-root"`, `aria-labelledby="scope-title external"`, `aria-describedby="scope-desc"`, `class="scope-one scope-two"`, `href="https://example.test/#title"`, `A &amp; B`} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in %s", want, got)
		}
	}
}

func TestSVGStaticNestedAnimations(t *testing.T) {
	source := `<svg id="root"><g><animate attributeName="opacity" values="0;.7;1"><desc>remove me</desc><!-- remove comment --></animate></g><path opacity="0"><animate attributeName="opacity" values="0;.5"/></path><path><animate attributeName="opacity" values="invalid;0;2;-1"/></path></svg>`
	got, ok := composeIconSVG(source, "scope", true)
	if !ok {
		t.Fatal("composition failed")
	}
	for _, unwanted := range []string{"<animate", "remove me", "remove comment", `opacity="0"`} {
		if strings.Contains(got, unwanted) {
			t.Errorf("static SVG retains %q", unwanted)
		}
	}
	for _, want := range []string{`<g opacity="1">`, `<path opacity=".5">`, `#scope-root,#scope-root *{animation:none!important;`} {
		if !strings.Contains(got, want) {
			t.Errorf("static SVG missing %q: %s", want, got)
		}
	}
}

func TestVisibleSVGOpacityBoundaries(t *testing.T) {
	for _, tc := range []struct{ input, want string }{
		{"", ""}, {"NaN;invalid;-1;0;1.1", ""}, {"0;.25; .75 ;.5", ".75"}, {"1;.5", "1"},
	} {
		if got := visibleSVGOpacity(tc.input); got != tc.want {
			t.Errorf("visibleSVGOpacity(%q)=%q, want %q", tc.input, got, tc.want)
		}
	}
}
