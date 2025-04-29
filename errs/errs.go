package errs

import (
	"log/slog"
	"net/http"

	"github.com/fivethirty/middest/handlers"
	"github.com/fivethirty/middest/internal/response"
)

type StatusError struct {
	error
	Status          int
	ResponseMessage string
}

func NewStatusError(
	err error,
	status int,
	responseMessage string,
) StatusError {
	return StatusError{
		error:           err,
		Status:          status,
		ResponseMessage: responseMessage,
	}
}

type HttpHandlerWithError func(http.ResponseWriter, *http.Request) error

type ErrorWriter func(w http.ResponseWriter, code int, responseMessage string)

type errs struct {
	writeError ErrorWriter
	logger     *slog.Logger
}

func New(
	writeError ErrorWriter,
	logger *slog.Logger,
) *errs {
	return &errs{
		writeError: writeError,
		logger:     logger,
	}
}

func (e *errs) WriteMissingErrorBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wrapped, ok := w.(*response.ResponseWriter)
		if !ok {
			wrapped = response.Wrap(w)
		}
		defer func() {
			if wrapped.IsHeaderWritten {
				return
			}
			e.writeError(wrapped, http.StatusInternalServerError, "")
		}()

		next.ServeHTTP(wrapped, r)

		if wrapped.IsHeaderWritten && wrapped.Status >= 400 && wrapped.BytesWritten == 0 {
			e.writeError(wrapped, wrapped.Status, "")
		}
	})
}

func (e *errs) ToHandlerFunc(he HttpHandlerWithError) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := he(w, r)
		if err == nil {
			return
		}

		e.logger.ErrorContext(
			r.Context(),
			"Handler Error",
			"message", err,
		)

		statusErr, ok := err.(StatusError)
		if !ok {
			e.writeError(w, http.StatusInternalServerError, "")
		} else {
			e.writeError(w, statusErr.Status, statusErr.ResponseMessage)
		}
	}
}

func (e *errs) WithMiddleware(
	middlewares []func(http.Handler) http.Handler,
	handlerWithError HttpHandlerWithError,
) http.HandlerFunc {
	return handlers.WithMiddleware(middlewares, e.ToHandlerFunc(handlerWithError))
}
