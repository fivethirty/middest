package errorbody

import (
	"net/http"

	"github.com/fivethirty/middest/internal/response"
)

type errorWriter func(w http.ResponseWriter, code int)

func New(ew errorWriter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wrapped, ok := w.(*response.ResponseWriter)
			if !ok {
				wrapped = response.Wrap(w)
			}
			defer func() {
				if wrapped.IsHeaderWritten {
					return
				}
				ew(wrapped, http.StatusInternalServerError)
			}()

			next.ServeHTTP(wrapped, r)

			if wrapped.IsHeaderWritten && wrapped.Status >= 400 && wrapped.BytesWritten == 0 {
				ew(wrapped, wrapped.Status)
			}
		})
	}
}
