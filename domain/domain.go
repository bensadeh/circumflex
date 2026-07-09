package domain

import (
	"net"
	"strings"

	"golang.org/x/net/idna"
	"golang.org/x/net/publicsuffix"
)

// FromURL returns the registrable domain (eTLD+1) of a URL for display:
// "https://blog.example.co.uk/post" yields "example.co.uk". It returns the
// empty string when no domain can be determined, such as for empty input,
// IP address hosts or a bare top-level domain. Punycode labels are
// converted to Unicode.
func FromURL(url string) string {
	host := hostname(url)
	if host == "" || net.ParseIP(host) != nil {
		return ""
	}

	if ascii, err := idna.ToASCII(host); err == nil {
		host = ascii
	}

	registrable, err := publicsuffix.EffectiveTLDPlusOne(host)
	if err != nil {
		return ""
	}

	if unicode, err := idna.ToUnicode(registrable); err == nil {
		return unicode
	}

	return registrable
}

// hostname extracts the lower-cased host from a URL that may lack a scheme,
// stripping credentials, port, path, query and fragment.
func hostname(url string) string {
	host := strings.ToLower(strings.TrimSpace(url))

	if index := strings.Index(host, "://"); index >= 0 {
		host = host[index+3:]
	}

	if index := strings.IndexAny(host, "/?#"); index >= 0 {
		host = host[:index]
	}

	if index := strings.LastIndexByte(host, '@'); index >= 0 {
		host = host[index+1:]
	}

	if strings.HasPrefix(host, "[") {
		end := strings.IndexByte(host, ']')
		if end < 0 {
			return ""
		}

		return host[1:end]
	}

	if index := strings.LastIndexByte(host, ':'); index >= 0 {
		host = host[:index]
	}

	return strings.Trim(host, ".")
}
