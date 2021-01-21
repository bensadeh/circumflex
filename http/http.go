package http

import (
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

func Get(url string) ([]byte, error) {
	client := &http.Client{Timeout: 2 * time.Second}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "circumflex")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer closeStream(resp.Body)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func closeStream(body io.ReadCloser) {
	_ = body.Close()
}
