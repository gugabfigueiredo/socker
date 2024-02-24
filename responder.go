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
	res *handlerResponse
	err *handlerError
	han http.HandlerFunc
}
