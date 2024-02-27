package socker

import (
	"io"
	"net/http"
	"strings"
)

type Requester struct {
	Method   string      `json:"method"`
	Path     string      `json:"path"`
	RawQuery string      `json:"query"`
	Headers  http.Header `json:"headers"`
	Body     string      `json:"body"`
}

func (r *Requester) parseBody(contentType string) (io.ReadCloser, error) {
	switch contentType {
	case "application/json":
		return io.NopCloser(strings.NewReader(r.Body)), nil
	default:
		return nil, nil
	}
}

func (r *Requester) ToHTTPRequest() (*http.Request, error) {
	parsedBody, err := r.parseBody(r.Headers.Get("Content-Type"))
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(r.Method, r.Path, parsedBody)
	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = r.RawQuery
	req.Header = r.Headers

	return req, nil
}
