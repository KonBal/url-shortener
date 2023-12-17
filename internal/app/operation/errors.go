package operation

import (
	"errors"
)

// Error Not Found.
var ErrNotFound error = errors.New("not found")

type notFoundError string

// Error returns string for error.
func (e notFoundError) Error() string {
	return string(e)
}

// Is checks that the target is Not Found.
func (e notFoundError) Is(target error) bool {
	return target == ErrNotFound
}

type notUniqueError struct {
	ShortURL string
}

// Error returns string for error.
func (e notUniqueError) Error() string {
	return "short url already exists"
}

// Error Deleted.
var ErrDeleted error = errors.New("deleted")

type deletedError string

// Error returns string for error.
func (e deletedError) Error() string {
	return string(e)
}

// Is checks that the target is Deleted.
func (e deletedError) Is(target error) bool {
	return target == ErrDeleted
}
