package socker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strings"
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

func (m *MockServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, pattern := range generalizePath(r.Method + " " + r.URL.Path) {
		if handler, ok := m.handlers[pattern]; ok {
			handler.ServeHTTP(w, r)
			return
		}
	}

	http.NotFound(w, r)
}

func (m *MockServer) Stop() {
	m.Server.Close()
}

type onSetting struct {
	On      string    `json:"on"`
	Path    string    `json:"path"`
	Request Requester `json:"request"`
	Handler Responder `json:"handler"`
}

func (m *MockServer) LoadFromFile(path string) error {
	// Open the JSON file
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Decode JSON from the file
	var data []onSetting
	err = json.NewDecoder(file).Decode(&data)
	if err != nil {
		return err
	}

	for _, setting := range data {
		switch setting.On {
		case "any", "ANY":
			m.OnAny(setting.Path).Respond(setting.Handler)
		case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch, http.MethodOptions, http.MethodHead, http.MethodConnect, http.MethodTrace:
			m.On(setting.On, setting.Path).Respond(setting.Handler)
		case "request", "REQUEST":
			req, err := setting.Request.ToRequest(m)
			if err != nil {
				return err
			}
			m.OnRequest(req).Respond(setting.Handler)
		default:
			return fmt.Errorf("unsupported method: %s", setting.On)
		}
	}

	return nil
}

func generalizePath(path string) []string {
	components := strings.Split(path, "/")
	patterns := make([]string, len(components))

	for i := range components {
		pattern := strings.Join(components[:i+1], "/")
		if i < len(components)-1 {
			pattern += "/*"
		}
		patterns[i] = pattern
	}

	sort.Sort(sort.Reverse(sort.StringSlice(patterns)))

	for i := 0; i < len(components); i++ {
		patterns = append(patterns, strings.TrimPrefix(patterns[i], components[0]))
	}

	return patterns
}
