package fetcher

import (
	"clx/http"
	"clx/types"
	"encoding/json"
)

const (
	submissionURL = "http://node-hnapi.herokuapp.com/news?page="
)

func FetchSubmissions(page string) ([]types.Submission, error) {
	JSON, err := http.Get(submissionURL + page)
	submissions := unmarshalJSON(JSON)
	return submissions, err
}

func unmarshalJSON(stream []byte) []types.Submission {
	var submissions []types.Submission
	_ = json.Unmarshal(stream, &submissions)
	return submissions
}
