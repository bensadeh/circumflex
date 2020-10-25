package fetcher

import (
	"clx/http"
	"clx/types"
	"encoding/json"
	"strconv"
)

const (
	submissionURL = "http://node-hnapi.herokuapp.com/news?page="
)

func FetchSubmissions(page int) ([]types.Submission, error) {
	p := strconv.Itoa(page)
	JSON, err := http.Get(submissionURL + p)
	submissions := unmarshalJSON(JSON)
	return submissions, err
}

func unmarshalJSON(stream []byte) []types.Submission {
	var submissions []types.Submission
	_ = json.Unmarshal(stream, &submissions)
	return submissions
}
