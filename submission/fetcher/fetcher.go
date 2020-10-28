package fetcher

import (
	"clx/http"
	"clx/types"
	"encoding/json"
	"strconv"
)

const (
	baseURL = "http://node-hnapi.herokuapp.com/"
	page = "?page="
)

func FetchSubmissions(page int, category int) ([]*types.Submission, error) {
	url := getUrl(category)
	p := strconv.Itoa(page)
	JSON, err := http.Get(url + p)
	submissions := unmarshalJSON(JSON)
	return submissions, err
}

func getUrl(category int) string {
	switch category {
	case types.News:
		return baseURL + "news" + page
	case types.Ask:
		return baseURL + "ask" + page
	default:
		panic("Category unsupported")
	}
}

func unmarshalJSON(stream []byte) []*types.Submission {
	var submissions []*types.Submission
	_ = json.Unmarshal(stream, &submissions)
	return submissions
}
