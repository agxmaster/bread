package httpcache

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"time"

	"git.qutoutiao.net/gopher/qms/pkg/qlog"
)

const (
	// XFromCache is the header added to responses that are returned from the cache
	XFromCache = "X-From-Cache"

	DefaultCacheSize = 1024
)

// A Cache interface is used by the Transport to store and retrieve responses.
type Cache interface {
	// Get returns the []byte representation of a cached response and a bool
	// set to true if the value isn't empty
	Get(key string) (responseBytes []byte, ok bool)
	// Set stores the []byte representation of a response against a key
	Set(key string, responseBytes []byte)
	// Delete removes the value associated with the key
	Delete(key string)
}

func NewCache(size int, expire time.Duration) Cache {
	cache, err := NewMemoryCache(size, expire)
	if err != nil {
		qlog.Errorf("NewMemoryCache error:%v", err)
		return &noopCache{}
	}
	return cache
}

// cacheKey returns the cache key for req.
func cacheKey(req *http.Request) string {
	if req.Method == http.MethodGet {
		return req.URL.String()
	}
	return req.Method + " " + req.URL.String()
}

// CachedResponse returns the cached http.Response for req if present, and nil
// otherwise.
func CachedResponse(c Cache, req *http.Request) (resp *http.Response, err error) {
	cachedVal, ok := c.Get(cacheKey(req))
	if !ok {
		return
	}

	b := bytes.NewBuffer(cachedVal)
	return http.ReadResponse(bufio.NewReader(b), req)
}

type Transport struct {
	// The RoundTripper interface actually used to make requests
	// If nil, http.DefaultTransport is used
	Transport http.RoundTripper
	Cache     Cache
	// If true, responses returned from the cache will be given an extra header, X-From-Cache
	MarkCachedResponses bool
}

// NewTransport returns a new Transport with the
// provided Cache implementation and MarkCachedResponses set to true
func NewTransport(c Cache) *Transport {
	return &Transport{Cache: c, MarkCachedResponses: true}
}

// Client returns an *http.Client that caches responses.
func (t *Transport) Client() *http.Client {
	return &http.Client{Transport: t}
}

// RoundTrip takes a Request and returns a Response
//
// If there is a fresh Response already in cache, then it will be returned without connecting to
// the server.
//
func (t *Transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	cacheKey := cacheKey(req)
	cacheable := (req.Method == "GET" || req.Method == "HEAD") && req.Header.Get("range") == ""

	var cachedResp *http.Response
	if cacheable {
		cachedResp, err = CachedResponse(t.Cache, req)
		if err == nil && cachedResp != nil {
			if t.MarkCachedResponses {
				cachedResp.Header.Set(XFromCache, "1")
			}
			return cachedResp, nil
		}
	}

	transport := t.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	resp, err = transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	if cacheable {
		switch req.Method {
		case "GET":
			// Delay caching until EOF is reached.
			resp.Body = &cachingReadCloser{
				R: resp.Body,
				OnEOF: func(r io.Reader) {
					resp := *resp
					resp.Body = ioutil.NopCloser(r)
					respBytes, err := httputil.DumpResponse(&resp, true)
					if err == nil {
						t.Cache.Set(cacheKey, respBytes)
					}
				},
			}
		default:
			respBytes, err := httputil.DumpResponse(resp, true)
			if err == nil {
				t.Cache.Set(cacheKey, respBytes)
			}
		}
	}
	return resp, nil
}

// cachingReadCloser is a wrapper around ReadCloser R that calls OnEOF
// handler with a full copy of the content read from R when EOF is
// reached.
type cachingReadCloser struct {
	// Underlying ReadCloser.
	R io.ReadCloser
	// OnEOF is called with a copy of the content of R when EOF is reached.
	OnEOF func(io.Reader)

	buf bytes.Buffer // buf stores a copy of the content of R.
}

// Read reads the next len(p) bytes from R or until R is drained. The
// return value n is the number of bytes read. If R has no data to
// return, err is io.EOF and OnEOF is called with a full copy of what
// has been read so far.
func (r *cachingReadCloser) Read(p []byte) (n int, err error) {
	n, err = r.R.Read(p)
	r.buf.Write(p[:n])
	if err == io.EOF {
		r.OnEOF(bytes.NewReader(r.buf.Bytes()))
	}
	return n, err
}

func (r *cachingReadCloser) Close() error {
	return r.R.Close()
}
