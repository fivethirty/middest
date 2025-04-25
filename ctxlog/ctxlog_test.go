package ctxlog_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/fivethirty/middest/ctxlog"
	"github.com/fivethirty/middest/internal/testhandler"
)

type logEntry struct {
	Level         string        `json:"level"`
	Message       string        `json:"msg"`
	Status        int           `json:"status"`
	Method        string        `json:"method"`
	Path          string        `json:"path"`
	Params        url.Values    `json:"params"`
	RequestID     string        `json:"request_id"`
	Duration      time.Duration `json:"duration"`
	ContentLength int           `json:"content_length"`
	UserID        string        `json:"user_id"`
}

func logs(buffer *bytes.Buffer, t *testing.T) []logEntry {
	t.Helper()
	var entries []logEntry
	dec := json.NewDecoder(buffer)
	for {
		var entry logEntry
		if err := dec.Decode(&entry); err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}
		entries = append(entries, entry)
	}
	return entries
}

func TestLog(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
		path   string
		status int
		size   int
		level  slog.Level
		userID string
	}{
		{
			name:   "OK should log INFO",
			method: http.MethodPost,
			path:   "/foo?bar=baz",
			status: http.StatusOK,
			size:   10,
			level:  slog.LevelInfo,
		},
		{
			name:   "redirect should log INFO",
			method: http.MethodPost,
			path:   "/foo?bar=baz",
			status: http.StatusFound,
			size:   0,
			level:  slog.LevelInfo,
		},
		{
			name:   "error should log ERROR",
			method: http.MethodPost,
			path:   "/foo?bar=baz",
			status: http.StatusNotFound,
			size:   5,
			level:  slog.LevelError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			buffer := bytes.NewBuffer(nil)
			logger := ctxlog.NewLogger(buffer)
			wrapped := ctxlog.New(logger)(testhandler.New(t, tt.status, tt.size))
			req := httptest.NewRequest(http.MethodPost, tt.path, nil)
			if tt.userID != "" {
				req.Context()
			}
			w := httptest.NewRecorder()
			wrapped.ServeHTTP(w, req)

			entries := logs(buffer, t)
			if len(entries) != 1 {
				t.Fatalf("expected 1 log entry, got %d", len(entries))
			}
			entry := entries[0]
			if entry.Level != tt.level.String() {
				t.Errorf("expected log level %s, got %s", tt.level.String(), entry.Level)
			}
			if entry.Method != tt.method {
				t.Errorf("expected method %s, got %s", tt.method, entry.Method)
			}
			u, err := url.Parse(tt.path)
			if err != nil {
				t.Fatal(err)
			}
			if entry.Path != u.Path {
				t.Errorf("expected path %s, got %s", tt.path, entry.Path)
			}
			if len(entry.Params) != len(u.Query()) {
				t.Errorf("expected %d params, got %d", len(u.Query()), len(entry.Params))
			}
			for k, v := range entry.Params {
				if v[0] != u.Query().Get(k) {
					t.Errorf("expected param %s=%s, got %s", k, u.Query().Get(k), v[0])
				}
			}
			if entry.RequestID == "" {
				t.Error("expected request ID to be set")
			}
			if entry.Duration == 0 {
				t.Error("expected duration to > 0")
			}
			if entry.ContentLength != tt.size {
				t.Errorf("expected content length %d, got %d", tt.size, entry.ContentLength)
			}
		})
	}
}

func TestLogsWithContext(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		attrs []slog.Attr
	}{
		{
			name: "no attributes",
		},
		{
			name: "attributes",
			attrs: []slog.Attr{
				slog.String("key", "value"),
				slog.Int("int", 101),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			buffer := bytes.NewBuffer(nil)
			ctx := context.Background()
			for _, attr := range tt.attrs {
				ctx = ctxlog.AppendCtx(ctx, attr)
			}

			logger := ctxlog.NewLogger(buffer)
			logger.Log(ctx, slog.LevelInfo, "test")
			var entry map[string]any
			dec := json.NewDecoder(buffer)
			if err := dec.Decode(&entry); err != nil {
				t.Fatal(err)
			}
			for _, attr := range tt.attrs {
				value, ok := entry[attr.Key]
				if !ok {
					t.Errorf("value for key %s not found in logs", attr.Key)
				}
				strValue := fmt.Sprintf("%v", value)
				if strValue != attr.Value.String() {
					t.Errorf("expected %v but got %v", attr.Value.Any(), strValue)
				}
			}
		})
	}
}
