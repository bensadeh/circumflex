package validator

import "strings"

func IsValidDomain(title, domain, category string) (bool, string) {
	if strings.Contains(category, "ask") {
		return false, "Reader Mode not supported for [purple]Ask HN[-] | Press Enter to go to the comment section"
	}

	if strings.Contains(domain, "twitter") ||
		strings.Contains(domain, "youtube") {
		return false, "Reader Mode not supported on this domain"
	}

	if strings.Contains(title, "[video]") {
		return false, "Reader Mode not supported for videos"
	}

	if strings.Contains(title, "[pdf]") {
		return false, "Reader Mode not supported for PDFs"
	}

	if strings.Contains(title, "[audio]") {
		return false, "Reader Mode not supported for audio"
	}

	return true, ""
}
