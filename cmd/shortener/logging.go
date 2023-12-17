package main

import (
	"net/http"
	"time"

	"github.com/KonBal/url-shortener/internal/app/logger"
)

type logHandler struct {
	log  *logger.Logger
	next http.Handler
}

// Created logging handler.
func LoggingHandler(logger *logger.Logger) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return &logHandler{
			log:  logger,
			next: h,
		}
	}
}

// ServeHTTP adds logging to pipeline.
func (h *logHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	responseData := &responseData{}
	rw := responseWriter{ResponseWriter: w, responseData: responseData}

	started := time.Now()
	h.next.ServeHTTP(&rw, req)
	duration := time.Since(started)

	h.log.Infoln(
		"uri", req.RequestURI,
		"method", req.Method,
		"duration", duration,
		"status", responseData.status,
		"size", responseData.size,
	)
}

type responseData struct {
	status int
	size   int
}

type responseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

// Write writes to response.
func (r *responseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

// WriteHeader writes header to response.
func (r *responseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}
