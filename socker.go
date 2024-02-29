package socker

import (
	"encoding/json"
	"fmt"
	"net"
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

func NewServerOnPort(port string) *MockServer {
	l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%s", port))
	if err != nil {
		panic(err)
	}

	m := &MockServer{
		handlers: make(map[string]*MockHandler),
	}

	m.Server = httptest.NewUnstartedServer(m)
	m.Server.Listener = l
	m.Server.Start()
	return m
}

func (m *MockServer) on(key string) *MockHandler {
	m.handlers[key] = &MockHandler{}
	return m.handlers[key]
}

func (m *MockServer) On(path string) *MockHandler {
	return m.on(makeHandlerKey(path))
}

func (m *MockServer) OnMethod(method, path string) *MockHandler {
	return m.on(makeHandlerKey(method, path))
}

func (m *MockServer) OnRequest(req *http.Request) *MockHandler {
	h := m.on(makeHandlerKey("REQUEST", req.Method, req.URL.Path))
	h.requester = req
	return h
}

func (m *MockServer) OnRoute(path string, router func(m *MockServer)) {
	sub := &MockServer{
		handlers: make(map[string]*MockHandler),
	}
	router(sub)
	for key, handler := range sub.handlers {
		routeKey := routeHandlerKey(path, key)
		m.handlers[routeKey] = handler
	}
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

type mockSetting struct {
	On      string    `json:"on"`
	Path    string    `json:"path"`
	Request Requester `json:"request"`
	Handler Responder `json:"handler"`
}

func (m *MockServer) LoadSettings(filePath string) error {
	// Open the JSON file
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Decode JSON from the file
	var data []mockSetting
	err = json.NewDecoder(file).Decode(&data)
	if err != nil {
		return err
	}

	for _, setting := range data {
		switch setting.On {
		case "any", "ANY":
			m.On(setting.Path).Respond(setting.Handler)
		case "method", "METHOD":
			m.OnMethod(setting.On, setting.Path).Respond(setting.Handler)
		case "request", "REQUEST":
			req, err := setting.Request.ToHTTPRequest()
			if err != nil {
				return err
			}
			m.OnRequest(req).Respond(setting.Handler)
		default:
			return fmt.Errorf("unsupported condition: %s", setting.On)
		}
	}

	return nil
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

func routeHandlerKey(route string, key string) string {
	parts := strings.Split(key, "/")

	head := parts[0]

	switch head {
	case "REQUEST":
		neck, tail := parts[1], parts[2:]
		return path.Join(append([]string{head, neck, route}, tail...)...)
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions, http.MethodHead, http.MethodConnect, http.MethodTrace:
		tail := parts[1:]
		return path.Join(append([]string{head, route}, tail...)...)
	default:
		return path.Join(route, key)
	}
}
