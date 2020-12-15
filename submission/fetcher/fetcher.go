package fetcher

import (
	"clx/http"
	"clx/types"
	"encoding/json"
	"strconv"
)

const (
	baseURL = "http://api.hackerwebapp.com/"
	page = "?page="
)

func FetchSubmissionEntries(page int, category int) ([]*types.Submission, error) {
	url := getUrl(category)
	p := strconv.Itoa(page)
	JSON, err := http.Get(url + p)
	submissions := unmarshalJSON(JSON)
	return submissions, err
}

func getUrl(category int) string {
	switch category {
	case types.FrontPage:
		return baseURL + "news" + page
	case types.New:
		return baseURL + "newest" + page
	case types.Ask:
		return baseURL + "ask" + page
	case types.Show:
		return baseURL + "show" + page
	default:
		panic("ApplicationState unsupported")
	}
}

func unmarshalJSON(stream []byte) []*types.Submission {
	var submissions []*types.Submission
	_ = json.Unmarshal(stream, &submissions)
	return submissions
}
