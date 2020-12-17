package fetcher

import (
	"clx/constants"
	"clx/http"
	"clx/structs"
	"encoding/json"
	"strconv"
)

const (
	baseURL = "http://api.hackerwebapp.com/"
	page = "?page="
)

func FetchSubmissionEntries(page int, category int) ([]*structs.Submission, error) {
	url := getUrl(category)
	p := strconv.Itoa(page)
	JSON, err := http.Get(url + p)
	submissions := unmarshalJSON(JSON)
	return submissions, err
}

func getUrl(category int) string {
	switch category {
	case constants.FrontPage:
		return baseURL + "news" + page
	case constants.New:
		return baseURL + "newest" + page
	case constants.Ask:
		return baseURL + "ask" + page
	case constants.Show:
		return baseURL + "show" + page
	default:
		panic("ApplicationState unsupported")
	}
}

func unmarshalJSON(stream []byte) []*structs.Submission {
	var submissions []*structs.Submission
	_ = json.Unmarshal(stream, &submissions)
	return submissions
}
