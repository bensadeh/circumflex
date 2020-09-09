package main

import (
	"io/ioutil"
	"net/http"
	"time"
)

type Comments struct {
	Name    string
	Body    string
	Comment []Comments
}

var myClient = &http.Client{Timeout: 10 * time.Second}

func getJSON(url string, target interface{}) ([]byte, error) {
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
