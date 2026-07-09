package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFromURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{"empty url", "", ""},
		{"plain domain", "example.com", "example.com"},
		{"scheme and path", "https://example.com/path/to/page", "example.com"},
		{"www is stripped", "https://www.example.com", "example.com"},
		{"subdomain is stripped", "https://blog.cloudflare.com/post", "cloudflare.com"},
		{"nested subdomains", "https://a.b.c.example.com", "example.com"},
		{"multi-part suffix", "https://www.bbc.co.uk/news", "bbc.co.uk"},
		{"subdomain with multi-part suffix", "https://foo.bar.co.uk", "bar.co.uk"},
		{"private suffix keeps owner label", "https://user.github.io/project", "user.github.io"},
		{"query without path", "https://example.com?page=2", "example.com"},
		{"fragment without path", "https://example.com#section", "example.com"},
		{"no scheme with path", "example.com/path", "example.com"},
		{"port is stripped", "https://example.com:8080/path", "example.com"},
		{"credentials are stripped", "https://user:pass@example.com/path", "example.com"},
		{"uppercase is lowered", "HTTPS://WWW.EXAMPLE.COM", "example.com"},
		{"trailing dot", "https://example.com./path", "example.com"},
		{"surrounding whitespace", " https://example.com ", "example.com"},
		{"punycode converts to unicode", "http://www.xn--bcher-kva.de/buch", "bücher.de"},
		{"unicode host", "https://bücher.de/buch", "bücher.de"},
		{"ipv4 host", "http://192.168.1.1/page", ""},
		{"ipv4 host with port", "http://192.168.1.1:8080/page", ""},
		{"ipv6 host", "http://[2001:db8::1]/page", ""},
		{"ipv6 host with port", "http://[2001:db8::1]:8080/page", ""},
		{"unclosed ipv6 bracket", "http://[2001:db8::1/page", ""},
		{"bare tld", "com", ""},
		{"scheme only", "https://", ""},
		{"hacker news item", "https://news.ycombinator.com/item?id=1", "ycombinator.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, FromURL(tt.url))
		})
	}
}
