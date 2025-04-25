package recoverer

import (
	"context"
	"log/slog"
	"net/http"
	"runtime/debug"
)

type logger interface {
	Log(ctx context.Context, level slog.Level, msg string, keysAndValues ...any)
}

func New(l logger) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					l.Log(
						r.Context(),
						slog.LevelError,
						"Panic",
						"error", err,
						"trace", string(debug.Stack()),
					)
				}
			}()
			h.ServeHTTP(w, r)
		})
	}
}
