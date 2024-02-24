package socker

import (
	"encoding/json"
	"io"
	"net/http"
	"time"
)

type MockHandler struct {
	path string
	mux  *http.ServeMux
	req  *http.Request
}

func (m *MockHandler) respond(r Responder) {
	m.mux.HandleFunc(m.path, func(w http.ResponseWriter, req *http.Request) {
		if !m.validateRequest(req) {
			http.Error(w, "Request does not match", http.StatusBadRequest)
			return
		}
		switch {
		case r.err != nil:
			http.Error(w, r.err.Message, r.err.Code)
			return
		case r.res != nil:
			w.Header().Set("Content-Type", r.res.contentType)
			w.WriteHeader(r.res.status)

			// Encode the body based on the specified mime-type
			if r.res.contentType == "" {
				return
			}

			switch r.res.contentType {
			case "application/json":
				if err := json.NewEncoder(w).Encode(r.res.body); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			default:
				// Default to string
				if _, err := io.WriteString(w, r.res.body.(string)); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			}
			return
		case r.han != nil:
			r.han(w, req)
			return
		}
	})
}

func (m *MockHandler) Respond(r Responder) {
	m.respond(r)
}

func (m *MockHandler) RespondError(code int, message string) {
	m.respond(Responder{
		err: &handlerError{
			Code:    code,
			Message: message,
		},
	})
}

func (m *MockHandler) RespondJSON(status int, body any) {
	m.respond(Responder{
		res: &handlerResponse{
			contentType: "application/json",
			status:      status,
			body:        body,
		},
	})
}

func (m *MockHandler) RespondStatus(status int) {
	m.respond(Responder{
		res: &handlerResponse{
			status: status,
		},
	})
}

func (m *MockHandler) RespondWith(handlerFunc http.HandlerFunc) {
	m.respond(Responder{
		han: handlerFunc,
	})
}

func (m *MockHandler) Timeout(delay time.Duration) {
	m.RespondWith(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(delay)
		w.WriteHeader(http.StatusOK)
	})
}

func (m *MockHandler) validateRequest(req *http.Request) bool {
	if m.req == nil {
		return true
	}

	if m.req.Method != req.Method || m.req.URL.String() != req.URL.String() {
		return false
	}

	// check if necessary headers are present
	for key, values1 := range m.req.Header {
		values2 := req.Header[key]

		if len(values1) != len(values2) {
			return false
		}
		for i := range values1 {
			if values1[i] != values2[i] {
				return false
			}
		}
	}

	// check if expected body is present
	switch {
	case m.req.Body == nil:
		return true
	case req.Body == nil:
		return false
	default:
		content1, err := io.ReadAll(m.req.Body)
		if err != nil {
			return false
		}

		content2, err := io.ReadAll(req.Body)
		if err != nil {
			return false
		}
		return string(content1) != string(content2)
	}
}
