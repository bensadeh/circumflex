package highlight

import "strings"

// HonorsDeclared reports whether a page's declared language should be trusted
// for this block. Most declarations pass through untouched; the exception is a
// trap label — highlight.js and Prism let an author stamp language-http on any
// block, so prose, shell and shortcodes reach the http lexer and render as a
// malformed request under a wrong "HTTP" heading. A declared http is honored
// only when the block opens with an HTTP start-line; the guesser (or the plain
// fallback) takes over otherwise.
func HonorsDeclared(text, lang string) bool {
	if resolvesToHTTP(lang) {
		return looksLikeHTTP(text)
	}

	return true
}

func resolvesToHTTP(lang string) bool {
	lexer := resolve(lang)

	return lexer != nil && lexer.Config().Name == "HTTP"
}

var httpMethods = map[string]struct{}{
	"GET": {}, "HEAD": {}, "POST": {}, "PUT": {}, "DELETE": {},
	"CONNECT": {}, "OPTIONS": {}, "TRACE": {}, "PATCH": {},
}

// looksLikeHTTP reports whether the first non-blank line is an HTTP message
// start-line: a request line (METHOD SP target SP HTTP/x.y) or a status line
// (HTTP/x.y SP nnn). The chroma lexer's root state matches one of these before
// it colors anything; without it the whole block lexes to error tokens, so the
// line's absence marks the http label as noise. The accepted versions mirror
// the lexer's own grammar so this honors exactly what chroma would color.
func looksLikeHTTP(text string) bool {
	var line string

	for l := range strings.SplitSeq(text, "\n") {
		if t := strings.TrimSpace(l); t != "" {
			line = t

			break
		}
	}

	fields := strings.Fields(line)
	if len(fields) < 2 {
		return false
	}

	if _, ok := httpMethods[fields[0]]; ok {
		return len(fields) == 3 && isHTTPVersion(fields[2])
	}

	return isHTTPVersion(fields[0]) && isStatusCode(fields[1])
}

// isHTTPVersion mirrors the lexer's HTTP/[123](?:\.[01])? version token.
func isHTTPVersion(s string) bool {
	rest, ok := strings.CutPrefix(s, "HTTP/")
	if !ok {
		return false
	}

	switch rest {
	case "1", "2", "3", "1.0", "1.1", "2.0", "2.1", "3.0", "3.1":
		return true
	default:
		return false
	}
}

func isStatusCode(s string) bool {
	if len(s) != 3 {
		return false
	}

	for i := range len(s) {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}

	return true
}
