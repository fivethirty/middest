package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fivethirty/middest/handlers"
)

func TestWithMiddleware(t *testing.T) {
	t.Parallel()

	executed := []int{}

	middleware := []func(next http.Handler) http.Handler{
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				executed = append(executed, 1)
				next.ServeHTTP(w, r)
			})
		},
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				executed = append(executed, 2)
				next.ServeHTTP(w, r)
			})
		},
	}

	handler := handlers.WithMiddleware(
		middleware,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if executed[0] != 1 || executed[1] != 2 {
		t.Fatalf("expected middleware order of execution to be 1, 2, got %v", executed)
	}
}
