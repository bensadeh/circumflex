package highlight

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// A declared http is a common highlight.js mislabel: honor it only when the
// block opens with an HTTP start-line, and never let it override a declaration
// that is not http.
func TestHonorsDeclared(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		text string
		lang string
		want bool
	}{
		{"http request line", "POST /wp-json/batch/v1 HTTP/1.1\nHost: example.com", "http", true},
		{"http request line http2", "GET / HTTP/2", "http", true},
		{"http status line", "HTTP/1.1 200 OK\nContent-Type: text/html", "http", true},
		{"http leading blank line", "\n\nGET /index.html HTTP/1.0", "http", true},

		{"http request trailing space", "GET /api HTTP/1.1 \nHost: x", "http", false},
		{"http request crlf line ending", "GET /api HTTP/1.1\r\nHost: x", "http", true},
		{"http request leading indent", "  GET /api HTTP/1.1\nHost: x", "http", true},
		{"http status trailing space", "HTTP/1.1 200 OK \nDate: now", "http", true},

		{"http prose", "Current task statement:", "http", false},
		{"http shortcode", "[embed]https://example.com[/embed]", "http", false},
		{"http bullet list", "- Check required and valid params\n- Dispatch each request", "http", false},
		{"http bare headers no start line", "Content-Type: application/json\nAuthorization: Bearer x", "http", false},
		{"http lowercase method", "post / HTTP/1.1", "http", false},
		{"http bad version", "GET / HTTP/9.9", "http", false},
		{"http method missing target", "GET HTTP/1.1", "http", false},

		{"non-trap language always honored", "Current task statement:", "php", true},
		{"non-trap json always honored", "not really json", "json", true},
		{"unknown lang honored", "whatever", "not-a-lexer", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, HonorsDeclared(tt.text, tt.lang))
		})
	}
}
