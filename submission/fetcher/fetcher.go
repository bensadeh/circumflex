package fetcher

import (
	"clx/constants/submissions"
	"clx/http"
	"clx/structs"
	"encoding/json"
	"strconv"
)

const (
	baseURL = "http://api.hackerwebapp.com/"
	page    = "?page="
)

func FetchSubmissionEntries(page int, category int) ([]*structs.Submission, error) {
	url := getUrl(category)
	p := strconv.Itoa(page)
	JSON, err := http.Get(url + p)
	subs := unmarshalJSON(JSON)
	return subs, err
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

func unmarshalJSON(stream []byte) []*structs.Submission {
	var subs []*structs.Submission
	_ = json.Unmarshal(stream, &subs)
	return subs
}
