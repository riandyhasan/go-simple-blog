package main

import (
	"encoding/json"
	"net/http"
)

type MiddlewareContext struct {
	ResponseWriter http.ResponseWriter
	Request        *http.Request
}

type AuthMiddlewareContext struct {
	ResponseWriter http.ResponseWriter
	Request        *http.Request
}

func (g *MiddlewareContext) ReturnError(status int, message string) error {
	g.ResponseWriter.WriteHeader(status)
	return json.NewEncoder(g.ResponseWriter).Encode(DefaultResponse{
		Status:  status,
		Message: message,
	})
}

func (g *MiddlewareContext) ReturnSuccess(data interface{}) error {
	g.ResponseWriter.WriteHeader(http.StatusOK)
	return json.NewEncoder(g.ResponseWriter).Encode(DefaultResponse{
		Status:  http.StatusOK,
		Message: "OK",
		Data:    data,
	})
}

func DefaultMiddleware(handlerFunc func(g *MiddlewareContext) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		guardCtx := MiddlewareContext{
			ResponseWriter: w,
			Request:        r,
		}
		if err := handlerFunc(&guardCtx); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
