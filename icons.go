package weather

import (
	"crypto/rand"
	"embed"
	"encoding/hex"
	"strconv"
	"sync/atomic"
)

//go:embed icons/*.svg
var iconFiles embed.FS

var (
	iconInstance     uint64
	iconProcessNonce = newIconNonce()
)

// GetIconSVG returns an SVG with unique IDs, classes, and animation names so
// separate calls can be inlined together, including across processes. A random
// process nonce and instance counter distinguish their names. Markup varies per call.
// Keys are exact lowercase icon names without paths or file extensions.
// Invalid or unknown keys return an empty string and false.
func GetIconSVG(key string) (svg string, ok bool) {
	return getIconSVG(key, nextIconPrefix(), false)
}

// GetIconSVGWithPrefix returns stable SVG markup for the given key and prefix.
// A prefix must start with an ASCII letter, contain only ASCII letters, digits,
// hyphens, or underscores, and have at most 64 characters. Use a different
// prefix for each instance in the same document, including repeated icons.
func GetIconSVGWithPrefix(key, prefix string) (svg string, ok bool) {
	if !validIconPrefix(prefix) {
		return "", false
	}
	// The length makes different caller prefixes occupy disjoint namespaces.
	return getIconSVG(key, "fcw-user-"+strconv.Itoa(len(prefix))+"-"+prefix, false)
}

// GetStaticIconSVG returns an SVG with unique instance names and no SMIL or CSS
// animation. It can be used when a viewer prefers reduced motion.
func GetStaticIconSVG(key string) (svg string, ok bool) {
	return getIconSVG(key, nextIconPrefix(), true)
}

func nextIconPrefix() string {
	return "fcw-auto-" + iconProcessNonce + "-" + strconv.FormatUint(atomic.AddUint64(&iconInstance, 1), 10)
}

func newIconNonce() string {
	var nonce [16]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		// Never fall back to a shared namespace if random generation fails.
		panic("weather: generate SVG instance nonce: " + err.Error())
	}
	return hex.EncodeToString(nonce[:])
}

func validIconPrefix(prefix string) bool {
	if len(prefix) == 0 || len(prefix) > 64 || !asciiLetter(prefix[0]) {
		return false
	}
	for i := range prefix {
		char := prefix[i]
		if !asciiLetter(char) && (char < '0' || char > '9') && char != '-' && char != '_' {
			return false
		}
	}
	return true
}

func getIconSVG(key, prefix string, static bool) (string, bool) {
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
	return composeIconSVG(string(data), prefix, static)
}
