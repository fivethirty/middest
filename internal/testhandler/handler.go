package testhandler

import (
	"net/http"
	"strings"
	"testing"
)

func New(t *testing.T, status int, responseSize int) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		if responseSize > 0 {
			_, err := w.Write([]byte(strings.Repeat("a", responseSize)))
			if err != nil {
				t.Fatal(err)
			}
		}
	})
}
