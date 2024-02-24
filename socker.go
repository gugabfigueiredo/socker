package socker

import (
	"net/http"
	"net/http/httptest"
)

type httpServer = *httptest.Server

type httpMuxer struct {
	mux   *http.ServeMux
	paths map[string]*http.ServeMux
}

type MockServer struct {
	httpServer
	any  *httpMuxer
	muxs map[string]*httpMuxer
}

func NewMockServer() *MockServer {
	anyMuxr := &httpMuxer{mux: http.NewServeMux(), paths: make(map[string]*http.ServeMux)}
	return &MockServer{
		any:  anyMuxr,
		muxs: map[string]*httpMuxer{"ANY": anyMuxr},
	}
}

func (m *MockServer) getHandler(method, path string) *http.ServeMux {
	muxr, ok := m.muxs[method]
	if !ok {
		mux := http.NewServeMux()
		m.muxs[method] = &httpMuxer{
			mux:   mux,
			paths: map[string]*http.ServeMux{path: mux},
		}
		muxr = m.muxs[method]
	} else {
		muxr.paths[path] = muxr.mux
	}

	return muxr.mux
}

func (m *MockServer) on(method, path string) *MockHandler {
	mux := m.getHandler(method, path)

	return &MockHandler{
		path: path,
		mux:  mux,
	}
}

func (m *MockServer) On(method, path string) *MockHandler {
	return m.on(method, path)
}

func (m *MockServer) OnAny(path string) *MockHandler {
	return m.on("ANY", path)
}

func (m *MockServer) OnRequest(req *http.Request) *MockHandler {
	mux := m.getHandler(req.Method, req.URL.Path)
	return &MockHandler{
		path: req.URL.Path,
		mux:  mux,
		req:  req,
	}
}

func (m *MockServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if method, ok := m.muxs[r.Method]; ok {
		if handler, ok := method.paths[r.URL.Path]; ok {
			handler.ServeHTTP(w, r)
			return
		}
	}
	if handler, ok := m.any.paths[r.URL.Path]; ok {
		handler.ServeHTTP(w, r)
		return
	}

	http.NotFound(w, r)
}

func (m *MockServer) Start() {
	m.httpServer = httptest.NewServer(m)
}

func (m *MockServer) Stop() {
	m.httpServer.Close()
}
