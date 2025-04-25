package response

import "net/http"

type ResponseWriter struct {
	http.ResponseWriter
	Status          int
	BytesWritten    int64
	IsHeaderWritten bool
}

func Wrap(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		Status:         http.StatusOK,
	}
}

func (rw *ResponseWriter) WriteHeader(code int) {
	if rw.IsHeaderWritten {
		return
	}

	rw.Status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.IsHeaderWritten = true
}

func (rw *ResponseWriter) Write(body []byte) (int, error) {
	bytesWritten, err := rw.ResponseWriter.Write(body)
	rw.BytesWritten += int64(bytesWritten)
	rw.IsHeaderWritten = true
	return bytesWritten, err
}
