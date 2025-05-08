package requestsize_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/fivethirty/middest/internal/testhandler"
	"github.com/fivethirty/middest/requestsize"
)

func TestRequestSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		maxBytes       int64
		requestSize    int
		contentLength  int
		expectedCode   int
		isBodyReadable bool
	}{
		{
			name:           "content length > max bytes, request size < max bytes should return 413",
			maxBytes:       10,
			requestSize:    5,
			contentLength:  15,
			expectedCode:   http.StatusRequestEntityTooLarge,
			isBodyReadable: true,
		},
		{
			name:           "content length and request size > bytes should return 413",
			maxBytes:       10,
			requestSize:    15,
			contentLength:  15,
			expectedCode:   http.StatusRequestEntityTooLarge,
			isBodyReadable: true,
		},
		{
			name:           "content length and request size < max bytes should return 200",
			maxBytes:       10,
			requestSize:    5,
			contentLength:  5,
			expectedCode:   http.StatusOK,
			isBodyReadable: true,
		},
		{
			name:           "content length < max bytes, request size > max bytes should not be readable",
			maxBytes:       10,
			requestSize:    15,
			contentLength:  5,
			expectedCode:   http.StatusOK,
			isBodyReadable: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			wrapped := requestsize.New(test.maxBytes)(testhandler.New(t, http.StatusOK, 0))
			body := strings.NewReader(strings.Repeat("a", int(test.requestSize)))
			req := httptest.NewRequest(http.MethodPost, "/", body)
			req.Header.Add("content-length", strconv.Itoa(test.contentLength))
			w := httptest.NewRecorder()
			wrapped.ServeHTTP(w, req)

			if w.Code != test.expectedCode {
				t.Errorf("expected status code %d, got %d", test.expectedCode, w.Code)
			}
			_, err := io.ReadAll(req.Body)
			_, ok := err.(*http.MaxBytesError)
			if !ok && err != nil {
				t.Fatal(err)
			}

			if ok && test.isBodyReadable {
				t.Errorf("expected request body to be readable but got error: %+v", err)
			} else if err == nil && !test.isBodyReadable {
				t.Errorf("expected request body to be unreadable")
			}
		})
	}
}
