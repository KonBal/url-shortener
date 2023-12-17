package logger

import (
	"net/http"

	"go.uber.org/zap"
)

// Custom logger over zap.SugaredLogger.
type Logger struct {
	zap.SugaredLogger
}

// NewLogger created new custom logger.
func NewLogger(logger *zap.Logger) *Logger {
	return &Logger{SugaredLogger: *logger.Sugar()}
}

// RequestError logs error err adding metadata of request req.
func (log *Logger) RequestError(req *http.Request, err error) {
	log.Errorf("%v %v: %v", req.Method, req.URL, err)
}
