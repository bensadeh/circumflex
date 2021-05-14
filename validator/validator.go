package validator

import "strings"

func GetErrorMessage(title, domain string) string {
	if strings.Contains(domain, "twitter") ||
		strings.Contains(domain, "youtube") ||
		strings.Contains(domain, "washingtonpost") ||
		strings.Contains(domain, "sciencedirect") ||
		strings.Contains(domain, "bloomberg.com") ||
		strings.Contains(domain, "drive.google.com") ||
		strings.Contains(domain, "spectrum.ieee.org") ||
		strings.Contains(domain, "marketplace.atlassian.com") ||
		strings.Contains(domain, "chrome.google.com") {
		return "Reader Mode not supported on " + domain
	}

	if strings.Contains(title, "[video]") {
		return "Reader Mode not supported for videos"
	}

	if strings.Contains(title, "[pdf]") {
		return "Reader Mode not supported for PDFs"
	}

	if strings.Contains(title, "[audio]") {
		return "Reader Mode not supported for audio"
	}

	if domain == "" {
		return "Reader Mode only supported on submissions with link"
	}

	return ""
}
