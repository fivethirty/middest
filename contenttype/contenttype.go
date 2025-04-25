package contenttype

import (
	"net/http"
	"slices"
)

func New(contentTypes ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength == 0 {
				next.ServeHTTP(w, r)
				return
			}
			contentType := r.Header.Get("content-type")
			if slices.Contains(contentTypes, contentType) {
				next.ServeHTTP(w, r)
				return
			}
			w.WriteHeader(http.StatusUnsupportedMediaType)
		})
	}
}
