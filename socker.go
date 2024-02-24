package socker

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
)

type MockServer struct {
	*httptest.Server
	any      *MockHandler
	handlers map[string]*MockHandler
}

func NewServer() *MockServer {
	m := &MockServer{
		handlers: make(map[string]*MockHandler),
	}
	m.Server = httptest.NewServer(m)
	return m
}

func (m *MockServer) on(key string) *MockHandler {
	m.handlers[key] = &MockHandler{}
	return m.handlers[key]
}

func (m *MockServer) On(method, path string) *MockHandler {
	return m.on(method + " " + path)
}

func (m *MockServer) OnAny(path string) *MockHandler {
	return m.on(path)
}

// OnRequest returns a validating handler for the given request method and path.
// If the incoming request does not match the given request, it returns a 400 Bad Request.
func (m *MockServer) OnRequest(req *http.Request) *MockHandler {
	h := m.on(req.Method + " " + req.URL.Path)
	h.req = req
	return h
}

// OnRequestStrict returns a handler strictly for the given request.
// This handler will only ever handle this exact same request including, including body.
// Other requests to the same endpoint are passed on and can be handled by any other matching handler.
func (m *MockServer) OnRequestStrict(req *http.Request) *MockHandler {
	h := m.on(hashRequest(req))
	h.req = req
	return h
}

func (m *MockServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if handler, ok := m.handlers[hashRequest(r)]; ok {
		handler.ServeHTTP(w, r)
		return
	}

	if handler, ok := m.handlers[r.Method+" "+r.URL.Path]; ok {
		handler.ServeHTTP(w, r)
		return
	}

	if handler, ok := m.handlers[r.URL.Path]; ok {
		handler.ServeHTTP(w, r)
		return
	}

	http.NotFound(w, r)
}

func (m *MockServer) Stop() {
	m.Server.Close()
}

func hashRequest(req *http.Request) string {
	// Read the request body and restore it
	body, err := io.ReadAll(req.Body)
	if err != nil {
		panic(fmt.Errorf("failed to read mock request body: %s", err))
	}
	req.Body = io.NopCloser(bytes.NewBuffer(body))

	// Extract relevant information from the request
	reqInfo := struct {
		Method string
		Path   string
		Query  url.Values
		Body   []byte
	}{
		Method: req.Method,
		Path:   req.URL.Path,
		Query:  req.URL.Query(),
		Body:   body,
	}

	// Sort headers for consistent hashing
	sortValuesMap(reqInfo.Query)

	// Convert request info to JSON
	reqJSON, err := json.Marshal(reqInfo)
	if err != nil {
		panic(fmt.Errorf("failed to marshal mock request info: %s", err))
	}

	// Hash the JSON-encoded request info
	hash := sha256.Sum256(reqJSON)
	return hex.EncodeToString(hash[:])
}

func sortValuesMap(m map[string][]string) {
	for key := range m {
		sorted := make([]string, len(m[key]))
		copy(sorted, m[key])
		sort.Strings(sorted)
		m[key] = sorted
	}
}
