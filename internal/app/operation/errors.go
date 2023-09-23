package operation

import (
	"errors"
	"net/http"
)

func logError(logger interface {
	Errorf(template string, args ...interface{})
}, req *http.Request, err error) {
	logger.Errorf("%v %v: %v", req.Method, req.URL, err)
}

var ErrNotFound error = errors.New("not found")

type notFoundError string

func (e notFoundError) Error() string {
	return string(e)
}

func (e notFoundError) Is(target error) bool {
	return target == ErrNotFound
}
