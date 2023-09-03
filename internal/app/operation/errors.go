package operation

import (
	"errors"
	"log"
	"net/http"
)

func logError(req *http.Request, err error) {
	log.Printf("%v %v: %v", req.Method, req.URL, err)
}

var ErrNotFound error = errors.New("not found")

type notFoundError string

func (e notFoundError) Error() string {
	return string(e)
}

func (e notFoundError) Is(target error) bool {
	return target == ErrNotFound
}
