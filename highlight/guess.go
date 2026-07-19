package highlight

import (
	"encoding/json"
	"path"
	"slices"
	"strings"
)

// GuessLang names the language of an unlabeled code block from structural
// signals alone, or returns "" when nothing matches confidently.
// Statistical classifiers were evaluated and rejected: on snippet-sized input
// go-enry's Bayes classifier names a wrong language with full confidence
// (Go → JavaScript, JSON → Python, logs → Markdown), and a wrongly colored
// block is worse than a plain one. Every detector here must stay silent on
// prose, logs, and ASCII diagrams.
func GuessLang(text string) string {
	if lang := shebangLang(text); lang != "" {
		return lang
	}

	lines := strings.Split(text, "\n")

	for _, d := range detectors {
		if d.match(text, lines) {
			return d.lang
		}
	}

	return ""
}

var detectors = []struct {
	lang  string
	match func(text string, lines []string) bool
}{
	{"diff", isDiff},
	{"json", isJSON},
	{"console", isShellSession},
	{"jsx", isComponentJSX},
	{"html", isHTML},
	{"xml", isXML},
	{"objective-c", isObjectiveC},
	{"cpp", isCPP},
	{"c", isC},
	{"sql", isSQL},
	{"docker", isDockerfile},
	{"php", isPHP},
	{"bash", isShell},
	{"go", isGo},
	{"rust", isRust},
	{"python", isPython},
	{"jsx", isJSX},
	{"javascript", isJavaScript},
}

// isJSX routes JavaScript with markup in expression position to the react
// lexer, which tags the JSX tags and attributes plain javascript leaves
// unstyled.
func isJSX(text string, lines []string) bool {
	return containsAny(text, []string{"(<", "=> <"}) && isJavaScript(text, lines)
}

// shebangLang maps an interpreter line to a lexer name; the strongest signal
// there is, when present.
func shebangLang(text string) string {
	first, _, _ := strings.Cut(text, "\n")
	if !strings.HasPrefix(first, "#!") {
		return ""
	}

	fields := strings.Fields(first[2:])
	if len(fields) == 0 {
		return ""
	}

	interpreter := path.Base(fields[0])
	if interpreter == "env" && len(fields) > 1 {
		interpreter = fields[1]
	}

	switch {
	case interpreter == "sh" || interpreter == "bash" || interpreter == "zsh" ||
		interpreter == "ksh" || interpreter == "dash":
		return "bash"

	case strings.HasPrefix(interpreter, "python"):
		return "python"

	case strings.HasPrefix(interpreter, "node"):
		return "javascript"

	case interpreter == "perl" || interpreter == "ruby" || interpreter == "awk" ||
		interpreter == "php" || interpreter == "lua":
		return interpreter

	default:
		return ""
	}
}

func isDiff(_ string, lines []string) bool {
	return anyLinePrefix(lines, "diff --git", "@@ -") ||
		(anyLinePrefix(lines, "--- ") && anyLinePrefix(lines, "+++ "))
}

// isJSON accepts only objects and arrays: bare strings and numbers are valid
// JSON too, but nothing worth coloring.
func isJSON(text string, _ []string) bool {
	trimmed := strings.TrimSpace(text)

	return (strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[")) &&
		json.Valid([]byte(trimmed))
}

// isShellSession keys on the "$ " prompt followed by something command-shaped
// — commands are lowercase or paths and never close with another $, which
// keeps dollar-delimited LaTeX out.
func isShellSession(_ string, lines []string) bool {
	for _, l := range lines {
		t := strings.TrimSpace(l)
		if !strings.HasPrefix(t, "$ ") || len(t) < 3 || strings.HasSuffix(t, "$") {
			continue
		}

		if c := t[2]; (c >= 'a' && c <= 'z') || c == '.' || c == '/' || c == '~' {
			return true
		}
	}

	return false
}

// isComponentJSX recognizes markup-first JSX: component tags are capitalized
// where html tags never are, and expression attributes use ={.
func isComponentJSX(text string, _ []string) bool {
	trimmed := strings.TrimSpace(text)

	return len(trimmed) >= 2 && trimmed[0] == '<' &&
		trimmed[1] >= 'A' && trimmed[1] <= 'Z' &&
		strings.Contains(trimmed, "={")
}

// isHTML claims markup only when the leading tag is an actual html element;
// everything else tag-shaped (pom files, Android layouts, SVG) is XML.
func isHTML(text string, _ []string) bool {
	trimmed := strings.TrimSpace(text)

	if strings.HasPrefix(strings.ToLower(trimmed), "<!doctype") {
		return true
	}

	if !markupShaped(trimmed) {
		return false
	}

	_, ok := htmlTags[strings.ToLower(leadingTag(trimmed))]

	return ok
}

func isXML(text string, _ []string) bool {
	trimmed := strings.TrimSpace(text)

	return strings.HasPrefix(trimmed, "<?xml") || markupShaped(trimmed)
}

// markupShaped wants a tag name after the bracket and either a closing tag
// or a self-closing element with an attribute — an angle-bracket autolink
// like <https://example.com/> has the slash but never the rest.
func markupShaped(trimmed string) bool {
	if len(trimmed) < 2 || trimmed[0] != '<' || !isASCIILetter(trimmed[1]) {
		return false
	}

	if leadingTag(trimmed) == "" {
		return false
	}

	return strings.Contains(trimmed, "</") ||
		(strings.Contains(trimmed, "/>") && strings.Contains(trimmed, "="))
}

// leadingTag returns the first tag's name, or "" when the bracket opens a
// URL scheme instead of an element.
func leadingTag(trimmed string) string {
	end := 1
	for end < len(trimmed) && (isASCIILetter(trimmed[end]) ||
		(trimmed[end] >= '0' && trimmed[end] <= '9') || trimmed[end] == '-') {
		end++
	}

	if end < len(trimmed) && trimmed[end] == ':' {
		return ""
	}

	return trimmed[1:end]
}

var htmlTags = map[string]struct{}{
	"html": {}, "head": {}, "body": {}, "div": {}, "span": {}, "p": {}, "a": {},
	"ul": {}, "ol": {}, "li": {}, "dl": {}, "dt": {}, "dd": {}, "table": {},
	"thead": {}, "tbody": {}, "tr": {}, "td": {}, "th": {}, "h1": {}, "h2": {},
	"h3": {}, "h4": {}, "h5": {}, "h6": {}, "img": {}, "br": {}, "hr": {},
	"form": {}, "input": {}, "button": {}, "label": {}, "select": {}, "option": {},
	"textarea": {}, "script": {}, "style": {}, "link": {}, "meta": {}, "title": {},
	"nav": {}, "header": {}, "footer": {}, "section": {}, "article": {}, "aside": {},
	"main": {}, "figure": {}, "figcaption": {}, "em": {}, "strong": {}, "b": {},
	"i": {}, "u": {}, "code": {}, "pre": {}, "blockquote": {}, "small": {},
	"sup": {}, "sub": {}, "iframe": {}, "video": {}, "audio": {}, "source": {},
	"canvas": {}, "details": {}, "summary": {},
}

func isASCIILetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isObjectiveC(text string, lines []string) bool {
	return anyLinePrefix(lines, "#import ") ||
		containsAny(text, []string{"@interface", "@implementation", "@property"})
}

func isCPP(text string, lines []string) bool {
	return hasInclude(lines) &&
		containsAny(text, []string{
			"std::", "<iostream>", "template<", "namespace ",
			"class ", "public:", "private:", "protected:", "virtual ", "nullptr",
		})
}

func isC(_ string, lines []string) bool {
	return hasInclude(lines)
}

func hasInclude(lines []string) bool {
	return anyLinePrefix(lines, "#include <", "#include \"")
}

// isSQL requires a leading SQL verb plus corroborating clauses — one for an
// all-caps verb, two otherwise, since lowercase from/where also read as
// English prose.
func isSQL(text string, lines []string) bool {
	var firstLine string

	for _, l := range lines {
		if t := strings.TrimSpace(l); t != "" {
			firstLine = t

			break
		}
	}

	first, _, _ := strings.Cut(firstLine, " ")
	verbs := []string{"select", "insert", "update", "delete", "create", "alter", "with", "explain"}

	if !slices.Contains(verbs, strings.ToLower(first)) {
		return false
	}

	// Prose defenses: statements don't end sentences with a period, don't
	// open with a colon clause ("UPDATE FAILED: …"), and aren't shouted
	// entirely in caps ("SELECT YOUR FAVORITE ITEMS FROM THE MENU");
	// lowercase SQL is additionally only trusted with a terminator.
	trimmed := strings.TrimSpace(text)
	if strings.HasSuffix(trimmed, ".") ||
		strings.Contains(firstLine, ": ") ||
		trimmed == strings.ToUpper(trimmed) {
		return false
	}

	needed := 2
	if first == strings.ToUpper(first) {
		needed = 1
	} else if !strings.Contains(text, ";") {
		return false
	}

	clauses := 0
	folded := " " + strings.ToLower(strings.Join(strings.Fields(text), " ")) + " "

	for _, c := range []string{" from ", " join ", " where ", " values", " table ", " group by ", " order by "} {
		if strings.Contains(folded, c) {
			clauses++
		}
	}

	return clauses >= needed
}

// isPHP takes the tag as decisive; without it, a $var assignment plus echo
// or $this-> names the language JavaScript's => threshold would otherwise
// claim.
func isPHP(text string, lines []string) bool {
	if anyLinePrefix(lines, "<?php") {
		return true
	}

	return atLeastTwo(
		dollarAssignment(lines),
		anyLinePrefix(lines, "echo ", "print "),
		strings.Contains(text, "$this->"),
	)
}

func isDockerfile(_ string, lines []string) bool {
	return anyLinePrefix(lines, "FROM ") &&
		anyLinePrefix(lines, "RUN ", "COPY ", "CMD ", "ENTRYPOINT ", "WORKDIR ", "ARG ", "ENV ")
}

func isShell(text string, lines []string) bool {
	// PHP and Perl assign to $vars; shell assigns without the sigil, so a
	// dollar-assignment line disqualifies the block outright.
	if dollarAssignment(lines) {
		return false
	}

	return atLeastTwo(
		anyLineIs(lines, "do", "done", "fi", "then", "esac", "elif"),
		dollarExpansion(text),
		containsAny(text, []string{">/dev/null", "2>&1"}),
		containsAny(text, []string{" | grep", " | awk", " | sed", " | sort", " | xargs", " | head", " | tail", " | wc"}),
		shellForIn(lines),
		anyLinePrefix(lines, "echo ", "cd ", "sudo ", "mkdir ", "curl ", "set -"),
		shellExport(lines),
		quotedExpansion(text) || containsAny(text, []string{"$#", "$?"}),
	)
}

// dollarExpansion counts $(cmd) substitutions and ${var} expansions —
// skipping Make's uppercase $(VAR) form and templating's ${{ doubles.
func dollarExpansion(text string) bool {
	for i := 0; i+2 < len(text); i++ {
		if text[i] != '$' {
			continue
		}

		c := text[i+2]

		switch text[i+1] {
		case '(':
			if c >= 'a' && c <= 'z' {
				return true
			}

		case '{':
			if c != '{' && (isASCIILetter(c) || c == '_' || (c >= '0' && c <= '9')) {
				return true
			}
		}
	}

	return false
}

// quotedExpansion looks for "$name — one "${ occurrence must not also feed
// this signal, or quoted template interpolation counts a single construct
// twice.
func quotedExpansion(text string) bool {
	for i := 0; i+2 < len(text); i++ {
		if text[i] == '"' && text[i+1] == '$' &&
			(isASCIILetter(text[i+2]) || text[i+2] == '_') {
			return true
		}
	}

	return false
}

func dollarAssignment(lines []string) bool {
	for _, l := range lines {
		t := strings.TrimSpace(l)
		if !strings.HasPrefix(t, "$") {
			continue
		}

		name := strings.TrimLeft(t[1:], "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_")
		if len(name) < len(t)-1 && strings.HasPrefix(strings.TrimSpace(name), "=") {
			return true
		}
	}

	return false
}

// shellExport separates shell's export VAR=value from JavaScript's export
// declarations, which always put spaces around any assignment.
func shellExport(lines []string) bool {
	for _, l := range lines {
		t := strings.TrimSpace(l)
		if strings.HasPrefix(t, "export ") && strings.Contains(t, "=") && !strings.Contains(t, " = ") {
			return true
		}
	}

	return false
}

func isGo(text string, lines []string) bool {
	return atLeastTwo(
		strings.Contains(text, " := "),
		anyLinePrefix(lines, "func "),
		anyLinePrefix(lines, "package "),
		containsAny(text, []string{"fmt.", "err != nil"}),
		anyLineIs(lines, "import ("),
	)
}

func isRust(text string, lines []string) bool {
	return atLeastTwo(
		anyLinePrefix(lines, "fn ", "pub fn "),
		containsAny(text, []string{"let mut ", "&mut ", "&str"}),
		unspacedPathSep(text),
		containsAny(text, []string{"println!", "#[derive", ".unwrap()", "?;", "vec!"}),
		anyLinePrefix(lines, "let "),
		rustMatchArm(text, lines),
	)
}

// unspacedPathSep reports Rust's Type::path form. Haskell's type-signature
// :: is always spaced, so spaced occurrences don't count.
func unspacedPathSep(text string) bool {
	for i := 0; ; i += 2 {
		j := strings.Index(text[i:], "::")
		if j < 0 {
			return false
		}

		i += j

		before := i == 0 || text[i-1] != ' '
		after := i+2 >= len(text) || text[i+2] != ' '

		if before && after {
			return true
		}
	}
}

// rustMatchArm pairs a match line opening a block with arrow arms, so match
// expressions read as Rust before the => alone can read as JavaScript.
func rustMatchArm(text string, lines []string) bool {
	if !strings.Contains(text, "=>") {
		return false
	}

	for _, l := range lines {
		t := strings.TrimSpace(l)
		if (strings.HasPrefix(t, "match ") || strings.Contains(t, " match ")) && strings.HasSuffix(t, "{") {
			return true
		}
	}

	return false
}

func isPython(text string, lines []string) bool {
	return atLeastTwo(
		pythonDef(lines),
		pythonImport(lines),
		containsAny(text, []string{"self.", "__init__", "__name__"}),
		containsAny(text, []string{"elif ", " is None", "f\""}),
		strings.Contains(text, "print("),
	)
}

func isJavaScript(text string, lines []string) bool {
	return atLeastTwo(
		strings.Contains(text, "=>"),
		anyLinePrefix(lines, "const ", "let ", "var "),
		containsAny(text, []string{"console.log", "===", "!=="}),
		containsAny(text, []string{"function ", "await ", "async "}),
		strings.Contains(text, "`${"),
		anyLinePrefix(lines, "export function", "export default", "export const",
			"export class", "export async", "export {", "import {"),
		containsAny(text, []string{"document.", "window.", ".addEventListener"}),
		containsAny(text, []string{"for (let ", "for (const ", "for (var "}),
		containsAny(text, []string{"(<", "=> <"}), // JSX flowing into an expression
	)
}

// pythonDef requires the trailing colon that separates a Python definition
// from Ruby's def or Go's func.
func pythonDef(lines []string) bool {
	for _, l := range lines {
		t := strings.TrimSpace(l)
		if (strings.HasPrefix(t, "def ") || strings.HasPrefix(t, "class ")) && strings.HasSuffix(t, ":") {
			return true
		}
	}

	return false
}

// pythonImport rejects lines with quotes (JavaScript's import-from-module
// form) and capitalized module paths — Python modules are lowercase where
// Swift imports Foundation and Java/Kotlin import java.util.Random.
func pythonImport(lines []string) bool {
	for _, l := range lines {
		t := strings.TrimSpace(l)
		if strings.ContainsAny(t, `"'`) {
			continue
		}

		var module string

		if rest, ok := strings.CutPrefix(t, "import "); ok {
			module, _, _ = strings.Cut(rest, " ")
		} else if rest, ok := strings.CutPrefix(t, "from "); ok && strings.Contains(t, " import ") {
			module, _, _ = strings.Cut(rest, " ")
		} else {
			continue
		}

		module = strings.TrimSuffix(module, ",")
		if module != "" && module == strings.ToLower(module) {
			return true
		}
	}

	return false
}

// shellForIn skips parenthesized loops: shell's for-in never has them,
// JavaScript's for (const key in obj) always does.
func shellForIn(lines []string) bool {
	for _, l := range lines {
		t := strings.TrimSpace(l)
		if strings.HasPrefix(t, "for ") && !strings.HasPrefix(t, "for (") &&
			strings.Contains(t, " in ") && !strings.HasSuffix(t, ":") {
			return true
		}
	}

	return false
}

func atLeastTwo(signals ...bool) bool {
	hits := 0

	for _, s := range signals {
		if s {
			hits++
		}
	}

	return hits >= 2
}

func anyLinePrefix(lines []string, prefixes ...string) bool {
	for _, l := range lines {
		t := strings.TrimSpace(l)

		for _, p := range prefixes {
			if strings.HasPrefix(t, p) {
				return true
			}
		}
	}

	return false
}

func anyLineIs(lines []string, words ...string) bool {
	for _, l := range lines {
		t := strings.TrimSpace(l)

		if slices.Contains(words, t) {
			return true
		}
	}

	return false
}

func containsAny(text string, targets []string) bool {
	for _, target := range targets {
		if strings.Contains(text, target) {
			return true
		}
	}

	return false
}
