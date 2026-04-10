package middleware

import (
	"context"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// RequestLogger returns a zerolog.Logger enriched with request context
// (request_id, user_id). Use this in handlers instead of the global logger
// to get correlation fields on every log line.
func RequestLogger(ctx context.Context) zerolog.Logger {
	l := log.With()
	if reqID := chimw.GetReqID(ctx); reqID != "" {
		l = l.Str("request_id", reqID)
	}
	if uid := GetUserID(ctx); uid != 0 {
		l = l.Int64("user_id", uid)
	}
	return l.Logger()
}
