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
		"apnews.com",
		"bloomberg.com",
		"chrome.google.com",
		"drive.google.com",
		"facebook.com",
		"jalopnik.com",
		"marketplace.atlassian.com",
		"newsweek.com",
		"npr.org",
		"old.reddit.com",
		"reddit.com",
		"sciencedirect.com",
		"security.googleblog.com",
		"spectrum.ieee.org",
		"twitter.com",
		"washingtonpost.com",
		"wsj.com",
		"youtube.com",
	}

	for _, invalidDomain := range invalidDomains {
		if domain == invalidDomain {
			return true
		}
	}

	return false
}
