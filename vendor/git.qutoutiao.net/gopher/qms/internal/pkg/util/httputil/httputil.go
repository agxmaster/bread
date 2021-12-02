package httputil

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"git.qutoutiao.net/gopher/qms/internal/core/common"
	"git.qutoutiao.net/gopher/qms/internal/core/invocation"
	"git.qutoutiao.net/gopher/qms/pkg/qlog"
)

//ErrInvalidReq invalid input
var ErrInvalidReq = errors.New("rest consumer call arg is not *http.Request type")

//SetURI sets host for the request.
//set http(s)://{domain}/xxx
func SetURI(req *http.Request, url string) {
	if tempURL, err := req.URL.Parse(url); err == nil {
		req.URL = tempURL
	}
}

//SetBody is a method used for setting body for a request
func SetBody(req *http.Request, body []byte) {
	req.Body = ioutil.NopCloser(bytes.NewReader(body))
}

//SetCookie set key value in request cookie
func SetCookie(req *http.Request, k, v string) {
	c := &http.Cookie{
		Name:  k,
		Value: v,
	}
	req.AddCookie(c)
}

//GetCookie is a method which gets cookie from a request
func GetCookie(req *http.Request, key string) string {
	cookie, err := req.Cookie(key)
	if err == http.ErrNoCookie {
		return ""
	}
	return cookie.Value
}

// SetContentType is a method used for setting content-type in a request
func SetContentType(req *http.Request, ct string) {
	req.Header.Set("Content-Type", ct)
}

// GetContentType is a method used for getting content-type in a request
func GetContentType(req *http.Request) string {
	return req.Header.Get("Content-Type")
}

//HTTPRequest convert invocation to http request
func HTTPRequest(inv *invocation.Invocation) (*http.Request, error) {
	reqSend, ok := inv.Args.(*http.Request)
	if !ok {
		return nil, ErrInvalidReq
	}
	// set header
	contextHeader := common.FromContext(inv.Ctx)
	reqSend.Header = MergeHttpHeader(reqSend.Header, contextHeader)
	return reqSend, nil
}

// ReadBody read body from the from the response
func ReadBody(resp *http.Response) []byte {
	if resp != nil && resp.Body != nil {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			qlog.Error(fmt.Sprintf("read body failed: %s", err.Error()))
			return nil
		}
		return body
	}
	qlog.Error("response body or response is nil")
	return nil
}

// GetRespCookie returns response Cookie.
func GetRespCookie(resp *http.Response, key string) []byte {
	for _, c := range resp.Cookies() {
		if c.Name == key {
			return []byte(c.Value)
		}
	}
	return nil
}

// SetRespCookie sets the cookie.
func SetRespCookie(resp *http.Response, cookie *http.Cookie) {
	resp.Header.Add("Set-Cookie", cookie.String())
}

func MergeHttpHeader(header http.Header, contextHeader common.Header) http.Header {
	h := contextHeader.Copy()
	for k, vv := range header {
		if len(h.Get(k)) == 0 {
			h.Set(k, vv...)
		}
	}
	return http.Header(h)
}
