package main

import "net/http"

type statusError struct {
	StatusCode int
	Err        error
}

func Unauthorized(err error) error {
	return statusError{
		StatusCode: http.StatusUnauthorized,
		Err:        err,
	}
}

func NotFound(err error) error {
	return statusError{
		StatusCode: http.StatusNotFound,
		Err:        err,
	}
}

func BadRequest(err error) error {
	return statusError{
		StatusCode: http.StatusBadRequest,
		Err:        err,
	}
}

func (err statusError) Error() string {
	return err.Err.Error()
}

func (err statusError) Unwrap() error {
	return err.Err
}

func (err statusError) Status() int {
	return err.StatusCode
}
