package weather

import "html/template"

// GetSVGCodeByName returns trusted, embedded SVG markup for an exact icon key.
// It returns an empty value for unknown keys. Each call uses the same instance
// isolation as GetIconSVG, so repeated icons can share an HTML document.
//
// Deprecated: Use GetIconSVG, which also reports whether the key exists.
func GetSVGCodeByName(name string) template.HTML {
	svg, ok := GetIconSVG(name)
	if !ok {
		return ""
	}
	// GetIconSVG only reads packaged assets; name is never used as markup.
	return template.HTML(svg)
}
