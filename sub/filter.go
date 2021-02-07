package sub

import "clx/core"

func Filter(submissions []*core.Submission, hideYCJobs bool) []*core.Submission {
	if !hideYCJobs {
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
