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

func (r *Requester) ToHTTPRequest() (*http.Request, error) {
	var body io.ReadCloser
	switch r.Headers.Get("Content-Type") {
	case "application/json":
		body = io.NopCloser(strings.NewReader(r.Body))
	}

	req, err := http.NewRequest(r.Method, r.Path, body)
	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = r.RawQuery
	req.Header = r.Headers

	return req, nil
}
