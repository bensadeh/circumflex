package filter

import "clx/core"

func Filter(stories []*core.Story, hideYCJobs bool) []*core.Story {
	if !hideYCJobs {
		return stories
	}

	var filtered []*core.Story

	for _, v := range stories {
		if v.Type == "job" {
			continue
		}

		filtered = append(filtered, v)
	}

	return filtered
}
