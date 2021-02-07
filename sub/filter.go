package sub

import "clx/core"

func Filter(submissions []*core.Submission, showIsHiring bool) []*core.Submission {
	if showIsHiring {
		return submissions
	}

	var filtered []*core.Submission

	for _, v := range submissions {
		if v.Type == "job" {
			continue
		}

		filtered = append(filtered, v)
	}

	return filtered
}
