package highlight

import "strings"

// isTOML wants a [section] header plus two typed bindings, at least one of
// them carrying a value only TOML writes that way — a quoted string, array
// or inline table. INI shares the section-and-equals skeleton but leaves its
// values bare, and git configs in articles are declared ini.
func isTOML(_ string, lines []string) bool {
	sections, bindings, typed := 0, 0, 0

	for _, l := range lines {
		t := strings.TrimSpace(l)

		if tomlSection(t) {
			sections++

			continue
		}

		if quoted, ok := tomlBinding(t); ok {
			bindings++

			if quoted {
				typed++
			}
		}
	}

	return sections >= 1 && bindings >= 2 && typed >= 1
}

func tomlSection(t string) bool {
	if len(t) < 3 || t[0] != '[' || !strings.HasSuffix(t, "]") {
		return false
	}

	inner := strings.Trim(t, "[]")

	return inner != "" && !strings.ContainsAny(inner, " (){}=")
}

// tomlBinding reports key = value with a bare key and a typed value; the
// second result is whether the value is quoted or structured, the forms INI
// never uses.
func tomlBinding(t string) (bool, bool) {
	name, value, ok := strings.Cut(t, " = ")
	if !ok || name == "" || value == "" || strings.HasSuffix(t, ";") {
		return false, false
	}

	if strings.TrimLeft(name, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-.\"'") != "" {
		return false, false
	}

	switch value[0] {
	case '"', '\'', '[', '{':
		return true, true
	}

	if (value[0] >= '0' && value[0] <= '9') || value[0] == '-' || value[0] == '+' ||
		value == "true" || value == "false" {
		return true, false
	}

	return false, false
}

// isCSS reads at-rules, selector-opening braces and known-property
// declarations. Sass constructs bail the whole block out: an scss file that
// looks exactly like CSS colors identically, but one with $variables or
// nesting would render wrong under the css lexer.
func isCSS(text string, lines []string) bool {
	if containsAny(text, []string{"@mixin", "@include", "@extend", "&:", "&."}) ||
		anyLinePrefix(lines, "$") {
		return false
	}

	// A stylesheet is punctuation-terminated nearly line for line, which a
	// larger document embedding a <style> section never is — the signals
	// below match its style lines, so the whole block must look the part.
	if !cssShaped(lines) {
		return false
	}

	return atLeastTwo(
		anyLinePrefix(lines, "@media", "@import ", "@keyframes ", "@font-face", ":root"),
		cssSelectorLine(lines),
		cssDeclarations(lines) >= 2,
	)
}

func cssShaped(lines []string) bool {
	var total, fits int

	for _, l := range lines {
		t := strings.TrimSpace(l)
		if t == "" {
			continue
		}

		total++

		switch {
		case strings.HasSuffix(t, "{"), strings.HasSuffix(t, "}"),
			strings.HasSuffix(t, ";"), strings.HasSuffix(t, ","),
			strings.HasPrefix(t, "/*"), strings.HasPrefix(t, "*"),
			strings.HasSuffix(t, "*/"), strings.HasPrefix(t, "@"):
			fits++
		}
	}

	return total > 0 && fits*10 >= total*9
}

// cssSelectorLine reports a line opening a rule: a class, id or known
// element selector before the brace. JavaScript's method chains also open
// lines with a dot and end with a brace, but their names precede a paren.
func cssSelectorLine(lines []string) bool {
	for _, l := range lines {
		t := strings.TrimSpace(l)
		if !strings.HasSuffix(t, "{") || strings.Contains(t, "(") {
			continue
		}

		if (t[0] == '.' || t[0] == '#') && len(t) > 1 && isASCIILetter(t[1]) {
			return true
		}

		first, _, _ := strings.Cut(t, " ")
		first, _, _ = strings.Cut(first, ":")
		first, _, _ = strings.Cut(first, ",")

		if _, ok := htmlTags[first]; ok {
			return true
		}
	}

	return false
}

func cssDeclarations(lines []string) int {
	count := 0

	for _, l := range lines {
		t := strings.TrimSpace(l)
		if !strings.HasSuffix(t, ";") {
			continue
		}

		prop, _, ok := strings.Cut(t, ": ")
		if !ok {
			continue
		}

		if _, exact := cssProperties[prop]; exact {
			count++

			continue
		}

		for _, family := range cssPropertyFamilies {
			if strings.HasPrefix(prop, family) {
				count++

				break
			}
		}
	}

	return count
}

var cssProperties = map[string]struct{}{
	"color": {}, "display": {}, "position": {}, "width": {}, "height": {},
	"top": {}, "left": {}, "right": {}, "bottom": {}, "opacity": {},
	"cursor": {}, "content": {}, "float": {}, "clear": {}, "z-index": {},
	"transform": {}, "transition": {}, "animation": {}, "gap": {},
	"line-height": {}, "letter-spacing": {}, "white-space": {}, "box-shadow": {},
	"align-items": {}, "justify-content": {}, "box-sizing": {}, "filter": {},
}

var cssPropertyFamilies = []string{
	"margin", "padding", "font", "background", "border", "text-", "flex",
	"grid", "overflow", "max-", "min-", "outline", "list-style",
}

// isMakefile pairs a column-zero target with its tab-indented recipe — the
// tab is the giveaway, YAML indents with spaces — and wants one more sign of
// make: automatic variables, .PHONY, or a conventional target name. A C
// goto label followed by tab-indented statements produces the pair alone.
func isMakefile(text string, lines []string) bool {
	return atLeastTwo(
		anyLinePrefix(lines, ".PHONY"),
		makeTargetRecipe(lines),
		containsAny(text, []string{"$@", "$<", "$^", "$(CC)", "$(MAKE)", "CFLAGS"}),
		makeConventionalTarget(lines),
	)
}

func makeTargetRecipe(lines []string) bool {
	for i, l := range lines {
		if !makeTargetLine(l) {
			continue
		}

		for _, next := range lines[i+1:] {
			if strings.TrimSpace(next) == "" {
				continue
			}

			if strings.HasPrefix(next, "\t") {
				return true
			}

			break
		}
	}

	return false
}

// makeTargetLine reports name: [deps] at column zero — no whitespace before
// the colon, so URLs, prose and YAML's indented keys don't shape up.
func makeTargetLine(l string) bool {
	if l == "" || l[0] == ' ' || l[0] == '\t' || l[0] == '#' {
		return false
	}

	name, _, ok := strings.Cut(l, ":")
	if !ok || name == "" {
		return false
	}

	return strings.TrimLeft(name, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-.%/") == ""
}

func makeConventionalTarget(lines []string) bool {
	for _, l := range lines {
		switch {
		case strings.HasPrefix(l, "all:"), strings.HasPrefix(l, "build:"),
			strings.HasPrefix(l, "clean:"), strings.HasPrefix(l, "install:"),
			strings.HasPrefix(l, "test:"), strings.HasPrefix(l, "run:"):
			return true
		}
	}

	return false
}

// isYAML closes the detector table: the colon-keyed line is the weakest
// shape here, so everything else gets to claim first. Two shapes must still
// agree — a document marker, a mapping in list position, repeated key-only
// openers, or a run of plain pairs — which chat transcripts and compiler
// logs (one kv shape, nothing else) never manage.
func isYAML(_ string, lines []string) bool {
	return atLeastTwo(
		yamlDocMarker(lines),
		yamlListMapping(lines),
		yamlKeyOpeners(lines) >= 2,
		yamlPlainPairs(lines) >= 2,
	) || yamlPairDensity(lines)
}

// yamlPairDensity accepts a flat config file: five pairs and next to
// nothing else. Compiler logs write the pair shape too, but through their
// severity words, and a glossary's values are sentences ending in periods —
// neither line counts, so neither block gets dense enough.
func yamlPairDensity(lines []string) bool {
	var total, pairs int

	for _, l := range lines {
		t := strings.TrimSpace(l)
		if t == "" || strings.HasPrefix(t, "#") {
			continue
		}

		total++

		key, value, found := strings.Cut(t, ": ")
		if !found || !yamlKey(key) || value == "" || strings.HasSuffix(t, ".") {
			continue
		}

		switch key {
		case "error", "warning", "note", "info", "debug", "hint", "fatal", "panic":
			continue
		}

		pairs++
	}

	return pairs >= 5 && pairs*10 >= total*9
}

func yamlDocMarker(lines []string) bool {
	for _, l := range lines {
		if strings.TrimSpace(l) == "---" {
			return true
		}
	}

	return false
}

// yamlListMapping reports - key: value, the list-of-mappings shape.
// A markdown bullet has no colon-terminated first word.
func yamlListMapping(lines []string) bool {
	for _, l := range lines {
		t := strings.TrimSpace(l)

		rest, ok := strings.CutPrefix(t, "- ")
		if !ok {
			continue
		}

		if key, _, found := strings.Cut(rest, ": "); found && yamlKey(key) {
			return true
		}

		if key, found := strings.CutSuffix(rest, ":"); found && yamlKey(key) {
			return true
		}
	}

	return false
}

// yamlKeyOpeners counts key: lines that open a nested mapping. Go's switch
// labels sit in this shape too, so the known statement keywords don't count.
func yamlKeyOpeners(lines []string) int {
	count := 0

	for _, l := range lines {
		t := strings.TrimSpace(l)

		key, found := strings.CutSuffix(t, ":")
		if !found || !yamlKey(key) {
			continue
		}

		switch key {
		case "default", "case", "try", "else", "finally", "public", "private", "protected":
			continue
		}

		count++
	}

	return count
}

func yamlPlainPairs(lines []string) int {
	count := 0

	for _, l := range lines {
		t := strings.TrimSpace(l)

		key, value, found := strings.Cut(t, ": ")
		if found && yamlKey(key) && value != "" &&
			!strings.HasSuffix(t, ";") && !strings.HasSuffix(t, ",") &&
			!strings.HasSuffix(t, "{") {
			count++
		}
	}

	return count
}

// yamlKey accepts bare lowercase keys — ident characters and dashes, the
// way workflow and compose files write them. An uppercase or quoted key
// exists in YAML but reads like prose labels (NOTE:, Warning:) too often.
func yamlKey(key string) bool {
	if key == "" || key[0] < 'a' || key[0] > 'z' {
		return false
	}

	return strings.TrimLeft(key, "abcdefghijklmnopqrstuvwxyz0123456789_-") == ""
}
