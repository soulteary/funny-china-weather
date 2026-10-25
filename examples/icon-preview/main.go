// Command icon-preview generates an offline gallery of the embedded SVG icons.
package main

import (
	_ "embed"
	"flag"
	"fmt"
	"html/template"
	"os"
	"sort"
	"strings"
	"time"

	weather "github.com/soulteary/funny-china-weather"
)

//go:embed page.html
var pageTemplate string

type icon struct {
	Key, Description, Codes, Period, Animation string
	Automatic                                  bool
	Animated, Static                           template.HTML
}

var manualDescriptions = map[string]string{
	"cloudy-day-variant-2":       "白天多云备用构图 2",
	"cloudy-day-variant-3":       "白天多云备用构图 3",
	"cloudy-night-variant-2":     "夜间多云备用构图 2",
	"cloudy-night-variant-3":     "夜间多云备用构图 3",
	"showers-day-variant-2":      "白天阵雨备用构图 2",
	"showers-day-variant-3":      "白天阵雨备用构图 3",
	"snow-showers-day-variant-2": "白天阵雪备用构图 2",
	"snow-showers-day-variant-3": "白天阵雪备用构图 3",
	"rain-variant-7":             "降雨备用构图 7",
	"lightning":                  "闪电与云的备用图形",
	"wind":                       "风；当前未按风力自动匹配",
}

func buildIcons(directory string) ([]icon, error) {
	// Derive automatic usage from the public selector, rather than duplicating
	// its mapping. Include unknown fallback and the extended supported codes.
	usage := map[string]map[int]string{}
	codes := []int{-1, 53, 301, 302}
	for code := 0; code <= 32; code++ {
		codes = append(codes, code)
	}
	for _, code := range codes {
		for _, hour := range []int{0, 12} {
			key, label := weather.GetWeatherIconAt(code, time.Date(2026, 10, 26, hour, 0, 0, 0, time.UTC))
			if usage[key] == nil {
				usage[key] = map[int]string{}
			}
			usage[key][code] = label
		}
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	var result []icon
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".svg") {
			continue
		}
		key := strings.TrimSuffix(entry.Name(), ".svg")
		svg, ok := weather.GetIconSVGWithPrefix(key, "preview-"+key)
		if !ok {
			return nil, fmt.Errorf("resource %q is not embedded", key)
		}
		static, ok := weather.GetStaticIconSVG(key)
		if !ok {
			return nil, fmt.Errorf("static resource %q is not embedded", key)
		}
		item := icon{Key: key, Animated: template.HTML(svg), Static: template.HTML(static), Period: "全天 / 通用"}
		if strings.Contains(key, "-day") {
			item.Period = "白天"
		}
		if strings.Contains(key, "-night") {
			item.Period = "夜间"
		}
		if strings.Contains(svg, "<animate") || strings.Contains(svg, "animation:") || strings.Contains(svg, "animation-name:") {
			item.Animation = "含动画"
		} else {
			item.Animation = "静态图形"
		}
		if matches := usage[key]; matches != nil {
			item.Automatic = true
			var numbers []int
			for code := range matches {
				numbers = append(numbers, code)
			}
			sort.Ints(numbers)
			var labels, values []string
			seen := map[string]bool{}
			for _, code := range numbers {
				values = append(values, fmt.Sprint(code))
				if !seen[matches[code]] {
					labels = append(labels, matches[code])
					seen[matches[code]] = true
				}
			}
			item.Description = strings.Join(labels, "、")
			item.Codes = strings.Join(values, ", ")
			if key == "unknown" {
				item.Codes = "未识别代码（包括 -1）"
			}
		} else {
			item.Description = manualDescriptions[key]
			if item.Description == "" {
				return nil, fmt.Errorf("missing description for %q", key)
			}
			item.Codes = "无自动映射"
		}
		result = append(result, item)
	}
	for key := range usage {
		found := false
		for _, item := range result {
			if item.Key == key {
				found = true
			}
		}
		if !found {
			return nil, fmt.Errorf("selected resource %q is absent from directory", key)
		}
	}
	return result, nil
}

func render(directory string) ([]byte, error) {
	icons, err := buildIcons(directory)
	if err != nil {
		return nil, err
	}
	automatic := 0
	for _, item := range icons {
		if item.Automatic {
			automatic++
		}
	}
	data := struct {
		Icons                    []icon
		Total, Automatic, Manual int
	}{icons, len(icons), automatic, len(icons) - automatic}
	tmpl, err := template.New("preview").Parse(pageTemplate)
	if err != nil {
		return nil, err
	}
	var out strings.Builder
	if err = tmpl.Execute(&out, data); err != nil {
		return nil, err
	}
	return []byte(out.String()), nil
}

func main() {
	directory := flag.String("icons", "icons", "source icon directory (inventory only)")
	output := flag.String("output", "docs/icon-preview.html", "output HTML path")
	flag.Parse()
	page, err := render(*directory)
	if err == nil {
		err = os.WriteFile(*output, page, 0644)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("Generated", *output)
}
