package weather

import "embed"

//go:embed icons/*.svg
var iconFiles embed.FS

// GetIconSVG returns the embedded SVG for an exact lowercase icon key, such as
// "cloudy-day". Keys must omit the file extension and any path components.
// It returns an empty string and false when the key is invalid or unknown.
func GetIconSVG(key string) (svg string, ok bool) {
	if key == "" {
		return "", false
	}
	for _, char := range key {
		if (char < 'a' || char > 'z') && (char < '0' || char > '9') && char != '-' {
			return "", false
		}
	}

	data, err := iconFiles.ReadFile("icons/" + key + ".svg")
	if err != nil {
		return "", false
	}
	return string(data), true
}
