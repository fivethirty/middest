package ctxlog

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/fivethirty/middest/internal/response"
)

func New(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			wrapped, ok := w.(*response.ResponseWriter)
			if !ok {
				wrapped = response.Wrap(w)
			}

			requestID, err := requestID()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			ctx := AppendCtx(r.Context(), slog.String("method", r.Method))
			ctx = AppendCtx(ctx, slog.String("path", r.URL.EscapedPath()))
			ctx = AppendCtx(ctx, slog.Any("params", r.URL.Query()))
			ctx = AppendCtx(ctx, slog.String("request_id", requestID))

			r = r.WithContext(ctx)

			next.ServeHTTP(wrapped, r)

			level := slog.LevelInfo
			if wrapped.Status >= 400 {
				level = slog.LevelError
			}
			logger.Log(
				ctx,
				level,
				"Request",
				"status", wrapped.Status,
				"duration", time.Since(start),
				"content_length", wrapped.BytesWritten,
			)
		})
	}
}

const requestIDLength = 32

func requestID() (string, error) {
	b := make([]byte, requestIDLength)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

type contextKey string

const slogFields contextKey = "slog_fields"

var DefaultLogger *slog.Logger = NewLogger(os.Stdout)

func NewLogger(w io.Writer) *slog.Logger {
	handler := &contextHandler{
		Handler: slog.NewJSONHandler(w, nil),
	}
	return slog.New(handler)
}

type contextHandler struct {
	slog.Handler
}

func (ch *contextHandler) Handle(ctx context.Context, r slog.Record) error {
	if attrs, ok := ctx.Value(slogFields).([]slog.Attr); ok {
		for _, v := range attrs {
			r.AddAttrs(v)
		}
	}
	return ch.Handler.Handle(ctx, r)
}

func AppendCtx(ctx context.Context, attr slog.Attr) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	if v, ok := ctx.Value(slogFields).([]slog.Attr); ok {
		v = append(v, attr)
		return context.WithValue(ctx, slogFields, v)
	}

	v := []slog.Attr{}
	v = append(v, attr)
	return context.WithValue(ctx, slogFields, v)
}
