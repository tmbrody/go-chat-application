package middleware

import (
	"context"
	"go-chat-application/config"
	"net/http"
)

func WithDB(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), config.ApiCfg.DB, config.ApiCfg.DB)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
