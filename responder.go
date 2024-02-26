package socker

import "net/http"

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
	Res  *HandlerResponse `json:"response"`
	Err  *HandlerError    `json:"error"`
	Func http.HandlerFunc
}
