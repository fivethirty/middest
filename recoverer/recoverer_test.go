package recoverer_test

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fivethirty/middest/internal/testhandler"
	"github.com/fivethirty/middest/recoverer"
)

type testLogger struct {
	logCount int
}

func (l *testLogger) Log(_ context.Context, _ slog.Level, _ string, _ ...any) {
	l.logCount++
}

func TestRecover(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		handler      http.Handler
		expectedLogs int
	}{
		{
			name:         "no panic should return 200",
			handler:      testhandler.New(t, http.StatusOK, 0),
			expectedLogs: 0,
		},
		{
			name: "panic should return 500",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic("oops!")
			}),
			expectedLogs: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			logger := &testLogger{}
			wrapped := recoverer.New(logger)(test.handler)
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()
			wrapped.ServeHTTP(w, req)

			if logger.logCount != test.expectedLogs {
				t.Errorf("expected %d logs, got %d", test.expectedLogs, logger.logCount)
			}
		})
	}
}
