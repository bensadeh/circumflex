package article

import (
	"errors"
	"strings"
)

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
		return errors.New("reader mode not supported for this domain")
	}

	if domain == "" {
		return errors.New("reader mode only supported on submissions with link")
	}

	return nil
}

func isInvalidDomain(domain string) bool {
	invalidDomains := [...]string{
		"bloomberg.com",
		"chrome.google.com",
		"drive.google.com",
		"facebook.com",
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
