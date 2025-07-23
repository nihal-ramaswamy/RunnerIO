package dto

import "net/http"

type HttpMiddleware func(http.ResponseWriter, *http.Request)

type ErrorMessage struct {
	Message string `json:"message"`
}
