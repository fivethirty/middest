package response_test

import (
	"net/http"
	"testing"

	"github.com/fivethirty/middest/internal/response"
)

const (
	fakeBytesWritten = 13
)

type fakeResponseWriter struct{}

func (f *fakeResponseWriter) Header() http.Header {
	return http.Header{}
}
func (f *fakeResponseWriter) Write([]byte) (int, error) {
	return fakeBytesWritten, nil
}
func (f *fakeResponseWriter) WriteHeader(int) {}

func TestResponseWriter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		fn       func(*response.ResponseWriter)
		expected response.ResponseWriter
	}{
		{
			name: "initial state",
			fn:   func(w *response.ResponseWriter) {},
			expected: response.ResponseWriter{
				Status:          200,
				BytesWritten:    0,
				IsHeaderWritten: false,
			},
		},
		{
			name: "write header",
			fn: func(w *response.ResponseWriter) {
				w.WriteHeader(404)
			},
			expected: response.ResponseWriter{
				Status:          404,
				BytesWritten:    0,
				IsHeaderWritten: true,
			},
		},
		{
			name: "write body",
			fn: func(w *response.ResponseWriter) {
				w.Write([]byte("Hello, World!"))
			},
			expected: response.ResponseWriter{
				Status:          200,
				BytesWritten:    fakeBytesWritten,
				IsHeaderWritten: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := response.Wrap(http.ResponseWriter(&fakeResponseWriter{}))
			tt.fn(w)
			if w.Status != tt.expected.Status {
				t.Errorf("expected status %d, got %d", tt.expected.Status, w.Status)
			}
			if w.BytesWritten != tt.expected.BytesWritten {
				t.Errorf("expected bytes written %d, got %d", tt.expected.BytesWritten, w.BytesWritten)
			}
			if w.IsHeaderWritten != tt.expected.IsHeaderWritten {
				t.Errorf("expected header written %v, got %v", tt.expected.IsHeaderWritten, w.IsHeaderWritten)
			}
		})
	}
}
