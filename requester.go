package socker

import (
	"bytes"
	"net/http"
)

type Requester struct {
	Method string              `json:"method"`
	URL    string              `json:"url"`
	Header map[string][]string `json:"headers"`
	Body   []byte              `json:"body"`
}

func (r *Requester) ToRequest(m *MockServer) (*http.Request, error) {
	req, err := http.NewRequest(r.Method, m.URL+r.URL, bytes.NewReader(r.Body))
	if err != nil {
		return nil, err
	}

	req.Header = r.Header
	return req, nil
}
