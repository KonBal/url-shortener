package operation

import (
	"errors"
)

var ErrNotFound error = errors.New("not found")

type notFoundError string

func (e notFoundError) Error() string {
	return string(e)
}

func (e notFoundError) Is(target error) bool {
	return target == ErrNotFound
}

type notUniqueError struct {
	ShortURL string
}

func (e notUniqueError) Error() string {
	return "short url already exists"
}
