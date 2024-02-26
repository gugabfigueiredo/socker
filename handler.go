package socker

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

type MockHandler struct {
	req *http.Request
	res Responder
}

func (m *MockHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if !m.validateRequest(req) {
		http.Error(w, "Request does not match", http.StatusBadRequest)
		return
	}
	switch {
	case m.res.Err != nil:
		http.Error(w, m.res.Err.Message, m.res.Err.Status)
		return
	case m.res.Res != nil:
		if m.res.Res.Header != nil {
			for key, values := range m.res.Res.Header {
				w.Header().Set(key, values[0])
			}
		}

		if m.res.Res.ContentType != "" {
			w.Header().Set("Content-Type", m.res.Res.ContentType)
		}
		w.WriteHeader(m.res.Res.Status)
		if m.res.Res.Body != nil {
			switch m.res.Res.ContentType {
			case "application/json":
				if err := json.NewEncoder(w).Encode(m.res.Res.Body); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			default:
				if _, err := w.Write(m.res.Res.Body.([]byte)); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			}
		}
		return
	case m.res.Func != nil:
		m.res.Func(w, req)
		return
	}
}

func (m *MockHandler) Respond(r Responder) {
	m.res = r
}

func (m *MockHandler) RespondError(code int, message string) {
	m.Respond(Responder{
		Err: &HandlerError{
			Status:  code,
			Message: message,
		},
	})
}

func (m *MockHandler) RespondJSON(status int, body any) {
	m.Respond(Responder{
		Res: &HandlerResponse{
			ContentType: "application/json",
			Status:      status,
			Body:        body,
		},
	})
}

func (m *MockHandler) RespondStatus(status int) {
	m.Respond(Responder{
		Res: &HandlerResponse{
			Status: status,
		},
	})
}

func (m *MockHandler) RespondWith(handlerFunc http.HandlerFunc) {
	m.Respond(Responder{
		Func: handlerFunc,
	})
}

func (m *MockHandler) Timeout(delay time.Duration) {
	m.RespondWith(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(delay)
		w.WriteHeader(http.StatusOK)
	})
}

func (m *MockHandler) validateRequest(incoming *http.Request) bool {
	if m.req == nil {
		return true
	}

	if m.req.URL.Path != incoming.URL.Path {
		return false
	}

	// check if necessary query parameters are present
	incomingQuery := incoming.URL.Query()
	for key, values := range m.req.URL.Query() {
		incomingValues := incomingQuery[key]

		if len(values) != len(incomingValues) {
			return false
		}
		for i := range values {
			if values[i] != incomingValues[i] {
				return false
			}
		}

	}

	// check if necessary headers are present
	for key, values := range m.req.Header {
		incomingValues := incoming.Header[key]

		if len(values) != len(incomingValues) {
			return false
		}
		for i := range values {
			if values[i] != incomingValues[i] {
				return false
			}
		}
	}

	// check if expected Body is present
	switch {
	case m.req.Body == nil:
		return true
	case incoming.Body == nil:
		return false
	default:
		body, err := io.ReadAll(m.req.Body)
		if err != nil {
			return false
		}

		incomingBody, err := io.ReadAll(incoming.Body)
		if err != nil {
			return false
		}
		incoming.Body = io.NopCloser(bytes.NewBuffer(incomingBody))
		return string(body) != string(incomingBody)
	}
}
