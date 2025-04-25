package requestsize

import (
	"net/http"
	"strconv"
)

func New(maxBytes int64) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			length, err := strconv.Atoi(r.Header.Get("content-length"))
			if err == nil && length > int(maxBytes) {
				w.WriteHeader(http.StatusRequestEntityTooLarge)
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			h.ServeHTTP(w, r)
		})
	}
}
