package fetcher

import (
	"clx/constants/submissions"
	"clx/core"
	"clx/http"
	"encoding/json"
	"strconv"
)

const (
	baseURL = "http://api.hackerwebapp.com/"
	page    = "?page="
)

func FetchSubmissionEntries(page int, category int) ([]*core.Submission, error) {
	url := getUrl(category)
	p := strconv.Itoa(page)
	JSON, httpErr := http.Get(url + p)
	if httpErr != nil {
		return nil, httpErr
	}
	subs, jsonErr := unmarshalJSON(JSON)
	if jsonErr != nil {
		return nil, jsonErr
	}
	return subs, nil
}

func getUrl(category int) string {
	switch category {
	case submissions.FrontPage:
		return baseURL + "news" + page
	case submissions.New:
		return baseURL + "newest" + page
	case submissions.Ask:
		return baseURL + "ask" + page
	case submissions.Show:
		return baseURL + "show" + page
	default:
		panic("ApplicationState unsupported")
	}
}

func unmarshalJSON(stream []byte) ([]*core.Submission, error) {
	var subs []*core.Submission
	err := json.Unmarshal(stream, &subs)
	return subs, err
}
