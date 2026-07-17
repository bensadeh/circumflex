package article

import (
	"errors"
	nurl "net/url"
	"path"
	"strings"
)

// UnsupportedDomainError is typed so the status bar can highlight the domain.
type UnsupportedDomainError struct {
	Domain string
}

func (e *UnsupportedDomainError) Error() string {
	return e.Domain + " does not support articles in reader mode"
}

func Validate(title, domain string) error {
	if strings.Contains(title, "[video]") {
		return errors.New("reader mode not supported for videos")
	}

	if strings.Contains(title, "[pdf]") || strings.Contains(title, "[PDF]") {
		return errors.New("reader mode not supported for PDFs")
	}

	if strings.Contains(title, "[audio]") {
		return errors.New("reader mode not supported for audio")
	}

	if isInvalidDomain(domain) {
		return &UnsupportedDomainError{Domain: domain}
	}

	if domain == "" {
		return errors.New("reader mode only supported on submissions with link")
	}

	return nil
}

// ValidateURL is Validate for a bare link followed from inside an article:
// there is no story title to carry [pdf]-style tags, so the scheme, the
// domain blocklist, and the path's file type stand in for it. The reader's
// link selector shows failing links muted and inert.
func ValidateURL(rawURL string) error {
	u, err := nurl.ParseRequestURI(rawURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return errors.New("not a readable link")
	}

	if domain := strings.TrimPrefix(u.Hostname(), "www."); isInvalidDomain(domain) {
		return &UnsupportedDomainError{Domain: domain}
	}

	// A link with a known full-text mirror (arXiv /pdf) is fetched as HTML,
	// so its own extension doesn't matter.
	if fullTextURL(u) != "" {
		return nil
	}

	// Rejecting on the extension fails fast: without it the whole file
	// downloads only for extraction to fail on the binary body.
	switch strings.ToLower(path.Ext(u.Path)) {
	case ".pdf":
		return errors.New("reader mode not supported for PDFs")

	case ".zip", ".gz", ".tgz", ".bin", ".exe", ".dmg", ".iso",
		".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg",
		".mp3", ".mp4", ".mov", ".webm":
		return errors.New("reader mode not supported for this file type")
	}

	return nil
}

func isInvalidDomain(domain string) bool {
	invalidDomains := [...]string{
		"bloomberg.com",
		"chrome.google.com",
		"drive.google.com",
		"facebook.com",
		"ft.com",
		"lttlabs.com",
		"marketplace.atlassian.com",
		"old.reddit.com",
		"play.google.com",
		"reddit.com",
		"twitter.com",
		"washingtonpost.com",
		"wsj.com",
		"xkcd.com",
		"youtube.com",
	}

	for _, invalidDomain := range invalidDomains {
		if domain == invalidDomain {
			return true
		}
	}

	return false
}
