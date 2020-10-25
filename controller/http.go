package controller

import (
	"clx/types"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	submissionURL = "http://node-hnapi.herokuapp.com/news?page="
)

var myClient = &http.Client{Timeout: 10 * time.Second}

func getSubmissions(page string) ([]types.Submission, error) {
	JSON, err := get(submissionURL + page)
	submissions := unmarshalJSON(JSON)
	return submissions, err
}

func get(url string) ([]byte, error) {
	r, err := myClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer closeStream(r.Body)

	body, readError := ioutil.ReadAll(r.Body)
	if readError != nil {
		return nil, readError
	}

	return body, nil
}

func closeStream(body io.ReadCloser) {
	_ = body.Close()
}

func unmarshalJSON(stream []byte) []types.Submission {
	var submissions []types.Submission
	_ = json.Unmarshal(stream, &submissions)
	return submissions
}
