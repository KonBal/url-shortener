package logging

import (
	"net/http"
	"time"
)

type logger interface {
	Infoln(args ...interface{})
}

type handler struct {
	log logger
}

func Handler(logger logger) func(http.HandlerFunc) http.HandlerFunc {
	h := &handler{log: logger}
	return h.withLogging
}

func (h *handler) withLogging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		responseData := &responseData{}
		rw := responseWriter{ResponseWriter: w, responseData: responseData}

		started := time.Now()
		next(&rw, r)
		duration := time.Since(started)

		h.log.Infoln(
			"uri", r.RequestURI,
			"method", r.Method,
			"duration", duration,
			"status", responseData.status,
			"size", responseData.size,
		)
	}
}

type responseData struct {
	status int
	size   int
}

type responseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

func (r *responseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *responseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}
