package weather

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
)

type svgInspection struct {
	ids            map[string]bool
	references     []string
	classes        map[string]bool
	keyframes      map[string]bool
	colors         []string
	styles         string
	animationCount int
	root           xml.StartElement
}

func inspectSVG(t *testing.T, source string) svgInspection {
	t.Helper()
	result := svgInspection{ids: make(map[string]bool), classes: make(map[string]bool), keyframes: make(map[string]bool)}
	decoder := xml.NewDecoder(strings.NewReader(source))
	inStyle := false
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("invalid SVG: %v", err)
		}
		switch token := token.(type) {
		case xml.StartElement:
			if result.root.Name.Local == "" {
				result.root = token
			}
			if isSVGAnimation(token.Name.Local) {
				result.animationCount++
			}
			inStyle = token.Name.Local == "style"
			result.inspectAttributes(t, token.Attr)
		case xml.EndElement:
			if token.Name.Local == "style" {
				inStyle = false
			}
		case xml.CharData:
			if inStyle {
				result.styles += string(token)
				for _, match := range svgURLReference.FindAllStringSubmatch(string(token), -1) {
					result.references = append(result.references, match[1])
				}
				for _, match := range svgKeyframeName.FindAllStringSubmatch(string(token), -1) {
					result.keyframes[match[1]] = true
				}
			}
		}
	}
	for _, reference := range result.references {
		if !result.ids[reference] {
			t.Errorf("unresolved local SVG reference %q", reference)
		}
	}
	return result
}

func TestSVGCompositionAllAssets(t *testing.T) {
	entries, err := iconFiles.ReadDir("icons")
	if err != nil {
		t.Fatal(err)
	}
	globalIDs, globalClasses, globalKeyframes := map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, entry := range entries {
		key := strings.TrimSuffix(entry.Name(), ".svg")
		source, err := iconFiles.ReadFile("icons/" + entry.Name())
		if err != nil {
			t.Fatal(err)
		}
		original := inspectSVG(t, string(source))
		for instance := 0; instance < 2; instance++ {
			svg, ok := GetIconSVG(key)
			if !ok {
				t.Fatalf("missing icon %s", key)
			}
			got := inspectSVG(t, svg)
			checkScopedSVGClasses(t, key, got.styles)
			if fmt.Sprint(got.colors) != fmt.Sprint(original.colors) {
				t.Errorf("%s colors changed", key)
			}
			if got.animationCount != original.animationCount {
				t.Errorf("%s lost SMIL animations", key)
			}
			if len(got.classes) != len(original.classes) || len(got.keyframes) != len(original.keyframes) {
				t.Errorf("%s CSS identifiers changed in count", key)
			}
			if len(got.ids) != len(original.ids)+1 {
				t.Errorf("%s IDs = %d, want %d source IDs plus root", key, len(got.ids), len(original.ids))
			}
			checkSVGViewBox(t, key, got.root)
			for _, pair := range []struct {
				name          string
				local, global map[string]bool
			}{{"ID", got.ids, globalIDs}, {"class", got.classes, globalClasses}, {"keyframe", got.keyframes, globalKeyframes}} {
				for name := range pair.local {
					if pair.global[name] {
						t.Errorf("%s %s %q conflicts with another instance", key, pair.name, name)
					}
					pair.global[name] = true
				}
			}
		}
	}
}

func TestGetIconSVGWithPrefix(t *testing.T) {
	first, ok := GetIconSVGWithPrefix("dust", "Card_A-1")
	second, again := GetIconSVGWithPrefix("dust", "Card_A-1")
	if !ok || !again || first != second {
		t.Fatal("same key and prefix must produce stable markup")
	}
	for _, prefix := range []string{"", "1card", "-card", "_card", "has space", "a/b", "a.b", "a:root", `a"`, "a<", "a\\", "a\n", "晴", strings.Repeat("a", 65)} {
		if svg, ok := GetIconSVGWithPrefix("clear-day", prefix); ok || svg != "" {
			t.Errorf("accepted invalid prefix %q", prefix)
		}
	}
	if _, ok := GetIconSVGWithPrefix("clear-day", strings.Repeat("a", 64)); !ok {
		t.Fatal("rejected 64-character prefix")
	}
	checkCustomSVGNamespaces(t)
	for _, key := range []string{"missing", "day.svg", "../day"} {
		if svg, ok := GetIconSVGWithPrefix(key, "card"); ok || svg != "" {
			t.Errorf("accepted invalid key %q", key)
		}
		if svg, ok := GetStaticIconSVG(key); ok || svg != "" {
			t.Errorf("accepted invalid static key %q", key)
		}
	}
	a, _ := GetIconSVGWithPrefix("clear-day", "Card")
	b, _ := GetIconSVGWithPrefix("clear-day", "Card-blur")
	idsA, idsB := inspectSVG(t, a).ids, inspectSVG(t, b).ids
	for id := range idsA {
		if idsB[id] {
			t.Errorf("distinct custom prefixes overlap: %s", id)
		}
	}
}

func TestSVGCompositionPreservesColorsAndReferences(t *testing.T) {
	source := `<?xml version="1.0"?><svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" viewBox="0 0 64 64"><!-- keep credits --><defs><linearGradient id="abc"/><path id="target"/></defs><style><![CDATA[/* preserve #abc */ @keyframes pulse {from {opacity:.5} to {opacity:1}} #abc,.pulse {fill:#abc;stroke:#12abcd;filter:url("#abc");animation-name:pulse}]]></style><g class="pulse" fill="#abc" style="fill:#abc;stroke:#12abcd;animation-name:pulse"/><use href="#target" xlink:href="#target"/><animate id="loop" begin="0s; loop.end+.42s" end="target.click" attributeName="opacity" values="0;1"/></svg>`
	svg, ok := composeIconSVG(source, "probe", false)
	if !ok {
		t.Fatal("composition failed")
	}
	got := inspectSVG(t, svg)
	if !strings.Contains(svg, `<?xml version="1.0"?>`) || !strings.Contains(svg, "<!-- keep credits -->") {
		t.Error("source declaration or comment was lost")
	}
	if !strings.Contains(got.styles, "#probe-abc,.probe-pulse") || !strings.Contains(got.styles, "animation-name:probe-pulse") {
		t.Error("CSS selectors or animation names were not rewritten")
	}
	for _, literal := range []string{"fill:#abc", "stroke:#12abcd", "/* preserve #abc */"} {
		if !strings.Contains(got.styles, literal) {
			t.Errorf("CSS literal changed: %s", literal)
		}
	}
	if !strings.Contains(svg, `style="fill:#abc;stroke:#12abcd;animation-name:probe-pulse"`) {
		t.Error("inline style changed colors or lost animation isolation")
	}
	if !strings.Contains(svg, "probe-loop.end+.42s") || !strings.Contains(svg, "probe-target.click") {
		t.Error("SMIL event references were not rewritten")
	}
	if !strings.Contains(svg, `href="#probe-target"`) || !strings.Contains(svg, `xlink:href="#probe-target"`) {
		t.Error("href references were not rewritten")
	}
	if _, ok := composeIconSVG(`<svg><g></svg>`, "probe", false); ok {
		t.Error("accepted invalid XML")
	}
}

func TestGetStaticIconSVG(t *testing.T) {
	entries, _ := iconFiles.ReadDir("icons")
	for _, entry := range entries {
		key := strings.TrimSuffix(entry.Name(), ".svg")
		svg, ok := GetStaticIconSVG(key)
		if !ok {
			t.Fatalf("missing static icon %s", key)
		}
		got := inspectSVG(t, svg)
		if strings.Contains(svg, `opacity="0"`) {
			t.Errorf("%s retains invisible precipitation in static output", key)
		}
		if got.animationCount != 0 {
			t.Errorf("%s retains %d SMIL animations", key, got.animationCount)
		}
		rootID := ""
		for _, attr := range got.root.Attr {
			if attr.Name.Local == "id" {
				rootID = attr.Value
			}
		}
		if !strings.Contains(got.styles, "#"+rootID+",#"+rootID+" *{animation:none!important;-webkit-animation:none!important}") {
			t.Errorf("%s CSS animations are not disabled within its root", key)
		}
	}
}

func TestGetIconSVGConcurrentInstances(t *testing.T) {
	const count = 64
	results := make(chan string, count)
	var workers sync.WaitGroup
	for index := 0; index < count; index++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			svg, ok := GetIconSVG("thunderstorm-hail")
			if !ok {
				results <- ""
				return
			}
			results <- svg
		}()
	}
	workers.Wait()
	close(results)
	ids := make(map[string]bool)
	for svg := range results {
		if svg == "" {
			t.Error("concurrent lookup failed")
			continue
		}
		for id := range inspectSVG(t, svg).ids {
			if ids[id] {
				t.Errorf("concurrent calls reused ID %s", id)
			}
			ids[id] = true
		}
	}
}

func TestCSSAnimationReferencesResolve(t *testing.T) {
	entries, _ := iconFiles.ReadDir("icons")
	declaration := regexp.MustCompile(`(?:^|[;{])\s*(?:-[a-z]+-)?animation-name\s*:\s*([A-Za-z_][A-Za-z0-9_-]*)`)
	for _, entry := range entries {
		svg, _ := GetIconSVG(strings.TrimSuffix(entry.Name(), ".svg"))
		got := inspectSVG(t, svg)
		for _, match := range declaration.FindAllStringSubmatch(got.styles, -1) {
			if !got.keyframes[match[1]] {
				t.Errorf("%s unresolved CSS keyframe %s", entry.Name(), match[1])
			}
		}
	}
}

func TestSVGInstancesAcrossProcesses(t *testing.T) {
	const probe = "FCW_SVG_PROCESS_PREFIX_PROBE"
	if os.Getenv(probe) == "1" {
		svg, ok := GetIconSVG("fog")
		if !ok {
			os.Exit(2)
		}
		fmt.Print(svg)
		os.Exit(0)
	}
	first, ok := GetIconSVG("fog")
	if !ok {
		t.Fatal("icon unavailable")
	}
	child := exec.Command(os.Args[0], "-test.run=^TestSVGInstancesAcrossProcesses$")
	child.Env = append(os.Environ(), probe+"=1")
	second, err := child.Output()
	if err != nil {
		t.Fatal(err)
	}
	parentIDs, childIDs := inspectSVG(t, first).ids, inspectSVG(t, string(second)).ids
	for id := range parentIDs {
		if childIDs[id] {
			t.Errorf("separate processes reused SVG ID %q", id)
		}
	}
}

func (result *svgInspection) inspectAttributes(t *testing.T, attrs []xml.Attr) {
	t.Helper()
	for _, attr := range attrs {
		switch attr.Name.Local {
		case "id":
			if result.ids[attr.Value] {
				t.Errorf("duplicate ID %q within an icon", attr.Value)
			}
			result.ids[attr.Value] = true
		case "class":
			for _, value := range strings.Fields(attr.Value) {
				result.classes[value] = true
			}
		case "href":
			if strings.HasPrefix(attr.Value, "#") {
				result.references = append(result.references, strings.TrimPrefix(attr.Value, "#"))
			}
		case "begin", "end":
			result.inspectTiming(attr.Value)
		case "fill", "stroke", "stop-color":
			if strings.HasPrefix(attr.Value, "#") {
				result.colors = append(result.colors, attr.Value)
			}
		}
		for _, match := range svgURLReference.FindAllStringSubmatch(attr.Value, -1) {
			result.references = append(result.references, match[1])
		}
	}
}

func (result *svgInspection) inspectTiming(value string) {
	for _, part := range strings.Split(value, ";") {
		part = strings.TrimSpace(part)
		if dot := strings.IndexByte(part, '.'); dot > 0 {
			id := part[:dot]
			if asciiLetter(id[0]) {
				result.references = append(result.references, id)
			}
		}
	}
}

func checkSVGViewBox(t *testing.T, key string, root xml.StartElement) {
	t.Helper()
	for _, attr := range root.Attr {
		if attr.Name.Local == "viewBox" {
			fields := strings.Fields(attr.Value)
			if len(fields) != 4 {
				t.Fatalf("%s invalid viewBox %q", key, attr.Value)
			}
			for _, field := range fields {
				if _, err := strconv.ParseFloat(field, 64); err != nil {
					t.Errorf("%s invalid viewBox coordinate %q", key, field)
				}
			}
		}
	}
}

func checkCustomSVGNamespaces(t *testing.T) {
	t.Helper()
	for _, prefix := range []string{"Card", "Card-blur", "auto-1"} {
		svg, ok := GetIconSVGWithPrefix("clear-day", prefix)
		if !ok {
			t.Fatal("valid prefix rejected")
		}
		fixed := inspectSVG(t, svg)
		dynamic, _ := GetIconSVG("clear-day")
		automatic := inspectSVG(t, dynamic)
		for id := range fixed.ids {
			if automatic.ids[id] {
				t.Errorf("custom and automatic ID namespaces overlap: %s", id)
			}
		}
	}
}

func checkScopedSVGClasses(t *testing.T, key, styles string) {
	t.Helper()
	for _, match := range svgClassName.FindAllStringSubmatch(styles, -1) {
		if !strings.HasPrefix(match[1], "fcw-auto-") {
			t.Errorf("%s unscoped CSS class selector %q", key, match[1])
		}
	}
}
