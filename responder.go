package socker

import "net/http"

type handlerResponse struct {
	contentType string
	status      int
	body        any
}

type handlerError struct {
	Message string
	Code    int
}

type Responder struct {
	Res *handlerResponse `json:"response"`
	Err *handlerError    `json:"error"`
	han http.HandlerFunc `json:"-"`
}
