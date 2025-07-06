package contenttype_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fivethirty/middest/contenttype"
	"github.com/fivethirty/middest/internal/testhandler"
)

func TestContentType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		allowedContentTypes []string
		contentType         string
		expectedCode        int
		hasBody             bool
	}{
		{
			name:                "invalid content type with request body should return 415",
			allowedContentTypes: []string{"application/x-www-form-urlencoded"},
			contentType:         "application/json",
			expectedCode:        http.StatusUnsupportedMediaType,
			hasBody:             true,
		},
		{
			name:                "invalid content type without request body should return 200",
			allowedContentTypes: []string{"application/x-www-form-urlencoded"},
			contentType:         "application/json",
			expectedCode:        http.StatusOK,
			hasBody:             false,
		},
		{
			name: "valid content type with request body should return 200",
			allowedContentTypes: []string{
				"application/json",
				"application/x-www-form-urlencoded",
			},
			contentType:  "application/json",
			expectedCode: http.StatusOK,
			hasBody:      true,
		},
		{
			name: "valid content type without request body should return 200",
			allowedContentTypes: []string{
				"application/json",
				"application/x-www-form-urlencoded",
			},
			contentType:  "application/json",
			expectedCode: http.StatusOK,
			hasBody:      false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			wrapped := contenttype.New(
				test.allowedContentTypes...)(
				testhandler.New(t, http.StatusOK, 0),
			)
			var body io.Reader
			if test.hasBody {
				body = strings.NewReader("body!")
			}
			req := httptest.NewRequest(http.MethodPost, "/", body)
			req.Header.Add("content-type", test.contentType)
			w := httptest.NewRecorder()
			wrapped.ServeHTTP(w, req)

			if w.Code != test.expectedCode {
				t.Errorf("expected status code %d, got %d", test.expectedCode, w.Code)
			}
		})
	}
}
