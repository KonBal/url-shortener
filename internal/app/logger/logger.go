package logger

import (
	"net/http"

	"go.uber.org/zap"
)

type Logger struct {
	zap.SugaredLogger
}

func NewLogger(logger *zap.Logger) *Logger {
	return &Logger{SugaredLogger: *logger.Sugar()}
}

func (log *Logger) RequestError(req *http.Request, err error) {
	log.Errorf("%v %v: %v", req.Method, req.URL, err)
}
