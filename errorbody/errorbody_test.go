package errorbody_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/fivethirty/middest/errorbody"
	"github.com/fivethirty/middest/internal/testhandler"
)

func writeBody(w http.ResponseWriter, code int) {
	w.WriteHeader(code)
	_, _ = w.Write([]byte(strconv.Itoa(code)))
}

func TestErrorBody(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		handler      http.Handler
		expectedCode int
		expetedBody  string
	}{
		{
			name:         "2xx with no body",
			handler:      testhandler.New(t, http.StatusOK, 0),
			expectedCode: http.StatusOK,
			expetedBody:  "",
		},
		{
			name:         "3xx with no body",
			handler:      testhandler.New(t, http.StatusFound, 0),
			expectedCode: http.StatusFound,
			expetedBody:  "",
		},
		{
			name:         "4xx with no body",
			handler:      testhandler.New(t, http.StatusBadRequest, 0),
			expectedCode: http.StatusBadRequest,
			expetedBody:  strconv.Itoa(http.StatusBadRequest),
		},
		{
			name:         "2xx with body",
			handler:      testhandler.New(t, http.StatusOK, 2),
			expectedCode: http.StatusOK,
			expetedBody:  "aa",
		},
		{
			name:         "3xx with body",
			handler:      testhandler.New(t, http.StatusFound, 3),
			expectedCode: http.StatusFound,
			expetedBody:  "aaa",
		},
		{
			name:         "4xx with body",
			handler:      testhandler.New(t, http.StatusBadRequest, 4),
			expectedCode: http.StatusBadRequest,
			expetedBody:  "aaaa",
		},
		{
			name: "panic with no header",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic("oops!")
			}),
			expectedCode: http.StatusInternalServerError,
			expetedBody:  strconv.Itoa(http.StatusInternalServerError),
		},
		{
			name: "panic with header",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadGateway)
				panic("oops!")
			}),
			expectedCode: http.StatusBadGateway,
			expetedBody:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			handler := func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					defer func() {
						_ = recover()
					}()
					next.ServeHTTP(w, r)
				})
			}(errorbody.New(writeBody)(tt.handler))
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("expected status code %d, got %d", tt.expectedCode, w.Code)
			}
			if w.Body.String() != tt.expetedBody {
				t.Errorf("expected body %q, got %q", tt.expetedBody, w.Body.String())
			}
		})
	}
}
