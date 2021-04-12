package filter

import (
	"clx/endpoints"
)

func Filter(stories []*endpoints.Story, hideYCJobs bool) []*endpoints.Story {
	if !hideYCJobs {
		return stories
	}

	var filtered []*endpoints.Story

	for _, v := range stories {
		if v.Type == "job" {
			continue
		}

		filtered = append(filtered, v)
	}

	return filtered
}
