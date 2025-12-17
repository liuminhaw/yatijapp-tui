package data

import (
	"errors"

	"github.com/liuminhaw/yatijapp-tui/internal/authclient"
)

type UnauthorizedApiDataErr struct {
	Status int
	Err    error
	Msg    string
}

func (e UnauthorizedApiDataErr) Error() string {
	return e.Err.Error()
}

type UnmatchedApiRespDataErr struct {
	Status int
	Err    error
	Msg    string
}

func (e UnmatchedApiRespDataErr) Error() string {
	return e.Err.Error()
}

type UnexpectedApiDataErr struct {
	Err error
	Msg string
}

func (e UnexpectedApiDataErr) Error() string {
	return e.Err.Error()
}

type NotFoundApiDataErr struct {
	Err error
	Msg string
}

func (e NotFoundApiDataErr) Error() string {
	return e.Err.Error()
}

// respErrorCheck checks the error returned from an API request and categorizes it.
// If the error is checked as an authentication error (e.g., invalid or missing token),
// it returns a LoadApiDataErr with relevant details. For any other unexpected errors,
// it returns an UnexpectedApiDataErr.
func respErrorCheck(err error, reqErrMsg string) error {
	if err == nil {
		panic("respErrorCheck called with nil error")
	}

	var authErr authclient.ErrInvalidToken
	if errors.As(err, &authErr) {
		return UnauthorizedApiDataErr{
			Status: authErr.Status,
			Err:    err,
			Msg:    err.Error(),
		}
	}

	var tokenErr authclient.ErrMissingToken
	if errors.As(err, &tokenErr) {
		return UnauthorizedApiDataErr{
			Err: err,
			Msg: err.Error(),
		}
	}

	return UnexpectedApiDataErr{
		Err: err,
		Msg: reqErrMsg,
	}
}
