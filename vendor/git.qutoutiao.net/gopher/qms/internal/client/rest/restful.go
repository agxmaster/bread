package rest

import (
	"io"
	"net/http"
)

//NewRequest is a function which creates new request
func NewRequest(method, urlStr string, body io.Reader) (*http.Request, error) {
	return http.NewRequest(method, urlStr, body)
}

// NewResponse is creating the object of response
func NewResponse() *http.Response {
	resp := &http.Response{
		Header: http.Header{},
	}
	return resp
}
