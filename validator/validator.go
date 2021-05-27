package validator

import "strings"

func GetErrorMessage(title, domain string) string {
	if strings.Contains(title, "[video]") {
		return "Reader Mode not supported for videos"
	}

	if strings.Contains(title, "[pdf]") {
		return "Reader Mode not supported for PDFs"
	}

	if strings.Contains(title, "[audio]") {
		return "Reader Mode not supported for audio"
	}

	if isInvalidDomain(domain) {
		return "Reader Mode not supported on " + domain
	}

	if domain == "" {
		return "Reader Mode only supported on submissions with link"
	}

	return ""
}

func isInvalidDomain(domain string) bool {
	invalidDomains := [...]string{
		"twitter.com",
		"youtube.com",
		"washingtonpost.com",
		"sciencedirect.com",
		"newsweek.com",
		"apnews.com",
		"npr.org",
		"security.googleblog.com",
		"facebook.com",
		"wsj.com",
		"bloomberg.com",
		"drive.google.com",
		"reddit.com",
		"old.reddit.com",
		"spectrum.ieee.org",
		"marketplace.atlassian.com",
		"chrome.google.com",
	}

	for _, invalidDomain := range invalidDomains {
		if domain == invalidDomain {
			return true
		}
	}

	return false
}
