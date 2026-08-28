package contenttype

import (
	"mime"
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
			mediaType, _, err := mime.ParseMediaType(contentType)
			if err != nil || !slices.Contains(contentTypes, mediaType) {
				w.WriteHeader(http.StatusUnsupportedMediaType)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
