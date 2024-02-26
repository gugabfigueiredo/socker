package socker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
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

func makeHandlerKey(parts ...string) string {
	return strings.TrimSpace(path.Join(parts...))
}

func makeHandlerKeys(method, path string) []string {
	key := makeHandlerKey(method, path)
	parts := strings.Split(key, "/")
	keys := make([]string, len(parts))

	for i := range parts {
		pattern := strings.Join(parts[:i+1], "/")
		if i < len(parts)-1 {
			pattern += "/*"
		}
		keys[i] = pattern
	}

	for _, k := range keys[:len(parts)] {
		keys = append(keys, "REQUEST/"+k)
	}

	sort.Sort(sort.Reverse(sort.StringSlice(keys)))

	for i := 0; i < len(parts); i++ {
		keys = append(keys, strings.TrimPrefix(keys[i], "REQUEST/"+parts[0]))
	}

	return keys
}

func (m *MockServer) On(path string) *MockHandler {
	return m.on(makeHandlerKey(path))
}

func (m *MockServer) OnMethod(method, path string) *MockHandler {
	return m.on(makeHandlerKey(method, path))
}

// OnRequest returns a validating handler for the given request method and path.
// If the incoming request does not match the given request, it returns a 400 Bad Request.
func (m *MockServer) OnRequest(req *http.Request) *MockHandler {
	h := m.on(makeHandlerKey("REQUEST", req.Method, req.URL.Path))
	h.req = req
	return h
}

func (m *MockServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, key := range makeHandlerKeys(r.Method, r.URL.Path) {
		if handler, ok := m.handlers[key]; ok && handler.validateRequest(r) {
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
			m.On(setting.Path).Respond(setting.Handler)
		case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch, http.MethodOptions, http.MethodHead, http.MethodConnect, http.MethodTrace:
			m.OnMethod(setting.On, setting.Path).Respond(setting.Handler)
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
