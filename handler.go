package socker

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

type HandlerResponse struct {
	ContentType string      `json:"content_type"`
	Status      int         `json:"status"`
	Header      http.Header `json:"header"`
	Body        any         `json:"body"`
}

type HandlerError struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

type Responder struct {
	Response    *HandlerResponse `json:"response"`
	Error       *HandlerError    `json:"error"`
	HandlerFunc http.HandlerFunc
}

type MockHandler struct {
	requester *http.Request
	responder Responder
}

func (m *MockHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch {
	case m.responder.Error != nil:
		http.Error(w, m.responder.Error.Message, m.responder.Error.Status)
		return
	case m.responder.Response != nil:
		if m.responder.Response.Header != nil {
			for key, values := range m.responder.Response.Header {
				w.Header().Set(key, values[0])
			}
		}

		if m.responder.Response.ContentType != "" {
			w.Header().Set("Content-Type", m.responder.Response.ContentType)
		}
		w.WriteHeader(m.responder.Response.Status)
		if m.responder.Response.Body != nil {
			switch m.responder.Response.ContentType {
			case "application/json":
				if err := json.NewEncoder(w).Encode(m.responder.Response.Body); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			default:
				if _, err := w.Write(m.responder.Response.Body.([]byte)); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			}
		}
		return
	case m.responder.HandlerFunc != nil:
		m.responder.HandlerFunc(w, req)
		return
	}
}

func (m *MockHandler) Respond(r Responder) {
	m.responder = r
}

func (m *MockHandler) RespondError(code int, message string) {
	m.Respond(Responder{
		Error: &HandlerError{
			Status:  code,
			Message: message,
		},
	})
}

func (m *MockHandler) RespondJSON(status int, body any) {
	m.Respond(Responder{
		Response: &HandlerResponse{
			ContentType: "application/json",
			Status:      status,
			Body:        body,
		},
	})
}

func (m *MockHandler) RespondStatus(status int) {
	m.Respond(Responder{
		Response: &HandlerResponse{
			Status: status,
		},
	})
}

func (m *MockHandler) RespondWith(handlerFunc http.HandlerFunc) {
	m.Respond(Responder{
		HandlerFunc: handlerFunc,
	})
}

func (m *MockHandler) Timeout(delay time.Duration) {
	m.RespondWith(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(delay)
		w.WriteHeader(http.StatusOK)
	})
}

func (m *MockHandler) validateRequest(incoming *http.Request) bool {
	if m.requester == nil {
		return true
	}

	// check if necessary query parameters are present
	incomingQuery := incoming.URL.Query()
	for key, values := range m.requester.URL.Query() {
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
	for key, values := range m.requester.Header {
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
	case m.requester.Body == nil:
		return true
	case incoming.Body == nil:
		return false
	default:
		body, err := io.ReadAll(m.requester.Body)
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
