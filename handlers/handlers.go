package handlers

import "net/http"

func WithMiddleware(
	middlewares []func(http.Handler) http.Handler,
	handler http.Handler,
) http.HandlerFunc {
	next := handler
	for i := len(middlewares) - 1; i >= 0; i-- {
		next = middlewares[i](next)
	}
	return next.ServeHTTP
}
