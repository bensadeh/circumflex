package main

import (
	"io/ioutil"
	"net/http"
	"time"
)

var myClient = &http.Client{Timeout: 10 * time.Second}

func getJSON(url string) ([]byte, error) {
	r, err := myClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	body, readError := ioutil.ReadAll(r.Body)
	if readError != nil {
		return nil, readError
	}

	return body, nil
}
