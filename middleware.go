package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

type MiddlewareContext struct {
	ResponseWriter http.ResponseWriter
	Request        *http.Request
}

type AuthMiddlewareContext struct {
	ResponseWriter http.ResponseWriter
	Request        *http.Request
	Claims         CustomClaims
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
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(DefaultResponse{
				Status:  http.StatusInternalServerError,
				Message: err.Error(),
			})
		}
	}
}

func AuthMiddleware(handlerFunc func(g *MiddlewareContext) error, roles []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(DefaultResponse{
				Status:  http.StatusUnauthorized,
				Message: "Tidak ada autentikasi",
			})
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(DefaultResponse{
				Status:  http.StatusUnauthorized,
				Message: "Format autentikasi salah",
			})
			return
		}

		token := tokenParts[1]
		claims, err := VerifyJWT(token)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(DefaultResponse{
				Status:  http.StatusUnauthorized,
				Message: "Token expired",
			})
			return
		}

		authorized := false
		for _, role := range roles {
			if claims.Role == role {
				authorized = true
				break
			}
		}

		if !authorized {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(DefaultResponse{
				Status:  http.StatusForbidden,
				Message: "Anda tidak memiliki akses",
			})
			return
		}
		guardCtx := MiddlewareContext{
			ResponseWriter: w,
			Request:        r,
		}
		if err := handlerFunc(&guardCtx); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(DefaultResponse{
				Status:  http.StatusInternalServerError,
				Message: err.Error(),
			})
		}
	}
}
