package http2tooling

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

var _ http.RoundTripper = &http2roundTripper{}

func RoundTripper(rt http.RoundTripper) http.RoundTripper {
	return &http2roundTripper{rt: rt}
}

// http2roundTripper makes requests http2 ready. See EnsureGetBodyMethod for details.
type http2roundTripper struct {
	rt http.RoundTripper
}

func (rt *http2roundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	if err := EnsureGetBodyMethod(request); err != nil {
		return nil, err
	}
	return rt.rt.RoundTrip(request)
}

// EnsureGetBodyMethod ensures GetBody is set on request.
// support for retrying GOAWAY was added in Go 1.8:
// https://github.com/golang/net/commit/8dab9293431a241c65a8e8d46bbaa0492ec94849
// And we need this GetBody to be set on request.
func EnsureGetBodyMethod(originalReq *http.Request) error {
	if originalReq.Body != nil && originalReq.GetBody == nil {
		buf, err := io.ReadAll(originalReq.Body)
		if err != nil {
			return fmt.Errorf("could not read request body: %w", err)
		}
		body := bytes.NewReader(buf)
		originalReq.GetBody = func() (io.ReadCloser, error) {
			if _, serr := body.Seek(0, 0); serr != nil {
				return nil, serr
			}
			return io.NopCloser(body), nil
		}
		originalReq.Body = io.NopCloser(body)
	}
	return nil
}
