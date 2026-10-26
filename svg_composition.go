package weather

import (
	"encoding/xml"
	"io"
	"regexp"
	"strconv"
	"strings"
)

var (
	svgURLReference = regexp.MustCompile(`url\(\s*["']?#([A-Za-z_][A-Za-z0-9_.:-]*)["']?\s*\)`)
	svgClassName    = regexp.MustCompile(`\.([A-Za-z_][A-Za-z0-9_-]*)`)
	svgKeyframeName = regexp.MustCompile(`@(?:-[a-z]+-)?keyframes\s+([A-Za-z_][A-Za-z0-9_-]*)`)
)

type svgNames struct {
	ids           map[string]string
	css           map[string]string
	staticOpacity map[int64]string
}

// collectSVGNames validates the trusted embedded XML before lexical rewriting.
func collectSVGNames(source, prefix string) (svgNames, bool) {
	names := svgNames{ids: make(map[string]string), css: make(map[string]string), staticOpacity: make(map[int64]string)}
	decoder := xml.NewDecoder(strings.NewReader(source))
	inStyle := false
	var positions []int64
	for {
		position := decoder.InputOffset()
		token, err := decoder.Token()
		if err == io.EOF {
			return names, true
		}
		if err != nil {
			return svgNames{}, false
		}
		switch token := token.(type) {
		case xml.StartElement:
			names.collectElement(token, prefix, positions)
			positions = append(positions, position)
			inStyle = token.Name.Local == "style"
		case xml.EndElement:
			positions = positions[:len(positions)-1]
			if token.Name.Local == "style" {
				inStyle = false
			}
		case xml.CharData:
			if inStyle {
				for _, match := range svgClassName.FindAllStringSubmatch(string(token), -1) {
					names.css[match[1]] = prefix + "-" + match[1]
				}
				for _, match := range svgKeyframeName.FindAllStringSubmatch(string(token), -1) {
					names.css[match[1]] = prefix + "-" + match[1]
				}
			}
		}
	}
}

func (names svgNames) collectElement(token xml.StartElement, prefix string, positions []int64) {
	if token.Name.Local == "animate" && len(positions) > 0 {
		attribute, values := "", ""
		for _, attr := range token.Attr {
			switch attr.Name.Local {
			case "attributeName":
				attribute = attr.Value
			case "values":
				values = attr.Value
			}
		}
		if attribute == "opacity" {
			if visible := visibleSVGOpacity(values); visible != "" {
				names.staticOpacity[positions[len(positions)-1]] = visible
			}
		}
	}
	for _, attr := range token.Attr {
		switch attr.Name.Local {
		case "id":
			names.ids[attr.Value] = prefix + "-" + attr.Value
		case "class":
			for _, name := range strings.Fields(attr.Value) {
				names.css[name] = prefix + "-" + name
			}
		}
	}
}

func composeIconSVG(source, prefix string, static bool) (string, bool) {
	names, ok := collectSVGNames(source, prefix)
	if !ok {
		return "", false
	}
	decoder := xml.NewDecoder(strings.NewReader(source))
	var out strings.Builder
	out.Grow(len(source) + 512)
	inStyle, rootSeen, skipDepth := false, false, 0
	for {
		start := decoder.InputOffset()
		token, err := decoder.RawToken()
		if err == io.EOF {
			return out.String(), true
		}
		if err != nil {
			return "", false
		}
		raw := source[start:decoder.InputOffset()]
		switch token := token.(type) {
		case xml.StartElement:
			if skipSVGElement(token.Name.Local, static, &skipDepth) {
				continue
			}
			root := !rootSeen
			rootSeen = true
			rootID := prepareSVGElement(&token, names, prefix, root, static, start)
			writeSVGStart(&out, token, strings.HasSuffix(raw, "/>"))
			inStyle = token.Name.Local == "style"
			if root && static {
				out.WriteString("<style><![CDATA[#" + rootID + ",#" + rootID + " *{animation:none!important;-webkit-animation:none!important}]]></style>")
			}
		case xml.EndElement:
			if skipDepth > 0 {
				skipDepth--
				continue
			}
			if token.Name.Local == "style" {
				inStyle = false
			}
			out.WriteString(raw)
		case xml.CharData:
			if skipDepth > 0 {
				continue
			}
			writeSVGText(&out, raw, string(token), names, inStyle)
		default:
			if skipDepth == 0 {
				out.WriteString(raw)
			}
		}
	}
}

func writeSVGStart(out *strings.Builder, token xml.StartElement, selfClosing bool) {
	out.WriteByte('<')
	writeSVGName(out, token.Name)
	for _, attr := range token.Attr {
		out.WriteByte(' ')
		writeSVGName(out, attr.Name)
		out.WriteString(`="`)
		_ = xml.EscapeText(out, []byte(attr.Value))
		out.WriteByte('"')
	}
	if selfClosing {
		out.WriteString("/>")
	} else {
		out.WriteByte('>')
	}
}

func writeSVGName(out *strings.Builder, name xml.Name) {
	if name.Space != "" {
		out.WriteString(name.Space)
		out.WriteByte(':')
	}
	out.WriteString(name.Local)
}

func rewriteSVGURLs(value string, ids map[string]string) string {
	return svgURLReference.ReplaceAllStringFunc(value, func(reference string) string {
		match := svgURLReference.FindStringSubmatch(reference)
		if replacement, ok := ids[match[1]]; ok {
			return "url(#" + replacement + ")"
		}
		return reference
	})
}

func rewriteSVGTiming(value string, ids map[string]string) string {
	parts := strings.Split(value, ";")
	for index, part := range parts {
		trimmed := strings.TrimSpace(part)
		if dot := strings.IndexByte(trimmed, '.'); dot > 0 {
			if replacement, ok := ids[trimmed[:dot]]; ok {
				parts[index] = strings.Replace(part, trimmed[:dot]+".", replacement+".", 1)
			}
		}
	}
	return strings.Join(parts, ";")
}

// The bundled stylesheets use ordinary rules and keyframe blocks. Tokenizing
// their identifiers avoids changing substrings, comments, quoted text, or hex
// colors while keeping classes and animation names in one instance namespace.
func rewriteSVGCSS(value string, names svgNames, inline bool) string {
	value = rewriteSVGURLs(value, names.ids)
	var out strings.Builder
	depth := 0
	if inline {
		depth = 1
	}
	for index := 0; index < len(value); {
		char := value[index]
		if end := svgCSSLiteralEnd(value, index); end > index {
			out.WriteString(value[index:end])
			index = end
			continue
		}
		if char == '#' {
			end := svgCSSIdentifierEnd(value, index+1)
			identifier := value[index+1 : end]
			if replacement, found := names.ids[identifier]; found && depth == 0 {
				out.WriteByte('#')
				out.WriteString(replacement)
			} else {
				out.WriteString(value[index:end])
			}
			index = end
			continue
		}
		if char == '{' {
			depth++
		}
		if char == '}' {
			depth--
		}
		if asciiLetter(char) || char == '_' || char == '-' {
			end := svgCSSIdentifierEnd(value, index+1)
			identifier := value[index:end]
			replacement, found := names.css[identifier]
			if found {
				out.WriteString(replacement)
			} else {
				out.WriteString(identifier)
			}
			index = end
			continue
		}
		out.WriteByte(char)
		index++
	}
	return out.String()
}

func asciiLetter(char byte) bool { return char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' }

func isSVGAnimation(name string) bool {
	return name == "animate" || name == "animateTransform" || name == "animateMotion" || name == "set"
}

func visibleSVGOpacity(values string) string {
	result, maximum := "", 0.0
	for _, value := range strings.Split(values, ";") {
		value = strings.TrimSpace(value)
		opacity, err := strconv.ParseFloat(value, 64)
		if err == nil && opacity > maximum && opacity <= 1 {
			result, maximum = value, opacity
		}
	}
	return result
}

func rewriteSVGAttribute(attr *xml.Attr, names svgNames) {
	switch attr.Name.Local {
	case "id":
		attr.Value = names.ids[attr.Value]
	case "class":
		classes := strings.Fields(attr.Value)
		for index, class := range classes {
			classes[index] = names.css[class]
		}
		attr.Value = strings.Join(classes, " ")
	case "href":
		if replacement, found := names.ids[strings.TrimPrefix(attr.Value, "#")]; found && strings.HasPrefix(attr.Value, "#") {
			attr.Value = "#" + replacement
		}
	case "begin", "end":
		attr.Value = rewriteSVGTiming(attr.Value, names.ids)
	case "aria-labelledby", "aria-describedby":
		fields := strings.Fields(attr.Value)
		for index, field := range fields {
			if replacement, found := names.ids[field]; found {
				fields[index] = replacement
			}
		}
		attr.Value = strings.Join(fields, " ")
	}
	attr.Value = rewriteSVGURLs(attr.Value, names.ids)
	if attr.Name.Local == "style" {
		attr.Value = rewriteSVGCSS(attr.Value, names, true)
	}
}

func prepareSVGElement(token *xml.StartElement, names svgNames, prefix string, root, static bool, start int64) string {
	rootID, hasID := prefix, false
	for index := range token.Attr {
		rewriteSVGAttribute(&token.Attr[index], names)
		if token.Attr[index].Name.Local == "id" {
			rootID, hasID = token.Attr[index].Value, true
		}
	}
	if static {
		if opacity := names.staticOpacity[start]; opacity != "" {
			setSVGOpacity(token, opacity)
		}
	}
	if root && !hasID {
		token.Attr = append(token.Attr, xml.Attr{Name: xml.Name{Local: "id"}, Value: rootID})
	}
	return rootID
}

// Preserve visible precipitation when its opacity animation is removed.
func setSVGOpacity(token *xml.StartElement, opacity string) {
	for index := range token.Attr {
		if token.Attr[index].Name.Local == "opacity" {
			token.Attr[index].Value = opacity
			return
		}
	}
	token.Attr = append(token.Attr, xml.Attr{Name: xml.Name{Local: "opacity"}, Value: opacity})
}

// Return the end of a comment or quoted string, leaving CSS literals untouched.
func svgCSSLiteralEnd(value string, index int) int {
	if strings.HasPrefix(value[index:], "/*") {
		end := strings.Index(value[index+2:], "*/")
		if end < 0 {
			return len(value)
		}
		return index + end + 4
	}
	char := value[index]
	if char != '\'' && char != '"' {
		return index
	}
	end := index + 1
	for end < len(value) {
		if value[end] == '\\' {
			end += 2
			continue
		}
		if value[end] == char {
			return end + 1
		}
		end++
	}
	return len(value)
}

func svgCSSIdentifierEnd(value string, index int) int {
	for index < len(value) && svgCSSIdentifierByte(value[index]) {
		index++
	}
	return index
}

func svgCSSIdentifierByte(char byte) bool {
	return asciiLetter(char) || char >= '0' && char <= '9' || char == '_' || char == '-'
}

func writeSVGText(out *strings.Builder, raw, value string, names svgNames, inStyle bool) {
	if inStyle {
		css := rewriteSVGCSS(value, names, false)
		out.WriteString("<![CDATA[" + strings.ReplaceAll(css, "]]>", "]]]]><![CDATA[>") + "]]>")
	} else {
		out.WriteString(raw)
	}
}

func skipSVGElement(name string, static bool, depth *int) bool {
	if *depth > 0 {
		(*depth)++
		return true
	}
	if static && isSVGAnimation(name) {
		*depth = 1
		return true
	}
	return false
}
