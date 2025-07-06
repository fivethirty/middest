package errs_test

import (
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fivethirty/middest/errs"
	"github.com/fivethirty/middest/internal/testhandler"
)

func writeError(w http.ResponseWriter, code int, responseMessage string) {
	w.WriteHeader(code)
	_, _ = w.Write([]byte(responseMessage))
}

func TestWriteMissingErrorBody(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		handler      http.Handler
		expectedCode int
		expetedBody  string
	}{
		{
			name:         "< 400 without body",
			handler:      testhandler.New(t, http.StatusOK, 0),
			expectedCode: http.StatusOK,
			expetedBody:  "",
		},
		{
			name:         "> 400 without body",
			handler:      testhandler.New(t, http.StatusBadRequest, 0),
			expectedCode: http.StatusBadRequest,
			expetedBody:  "",
		},
		{
			name:         "< 400 with body",
			handler:      testhandler.New(t, http.StatusOK, 2),
			expectedCode: http.StatusOK,
			expetedBody:  "aa",
		},
		{
			name:         "> 400 with body",
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
			expetedBody:  "",
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
		{
			name:         "no header results in 500",
			handler:      http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
			expectedCode: http.StatusInternalServerError,
			expetedBody:  "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			errs := errs.New(writeError, slog.Default())
			handler := func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					defer func() {
						_ = recover()
					}()
					next.ServeHTTP(w, r)
				})
			}(errs.WriteMissingErrorBody(test.handler))
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != test.expectedCode {
				t.Errorf("expected status code %d, got %d", test.expectedCode, w.Code)
			}
			if w.Body.String() != test.expetedBody {
				t.Errorf("expected body %q, got %q", test.expetedBody, w.Body.String())
			}
		})
	}
}

func TestToHandlerFunc(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		err            error
		expectedStatus int
	}{
		{
			name:           "should return 500 when error is not StatusError",
			err:            errors.New("some error"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "should return 400 when StatusError with status 400",
			err: errs.NewStatusError(
				errors.New("bad request"),
				http.StatusBadRequest,
				"bad request body",
			),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "should do nothing when error is nil",
			err:            nil,
			expectedStatus: http.StatusOK,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			e := errs.New(writeError, slog.Default())
			handler := e.ToHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
				return test.err
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			handler(w, req)

			if w.Code != test.expectedStatus {
				t.Errorf("expected status %d, got %d", test.expectedStatus, w.Code)
			}

			statusErr, isStatusErr := test.err.(errs.StatusError)
			if isStatusErr && statusErr.ResponseMessage != "" &&
				w.Body.String() != statusErr.ResponseMessage {
				t.Errorf(
					"expected response message %q, got %q",
					statusErr.ResponseMessage,
					w.Body.String(),
				)
			} else if !isStatusErr && statusErr.ResponseMessage != "" {
				t.Errorf("expected response message to be empty, got %q", w.Body.String())
			}
		})
	}
}
