package client

import (
	"fmt"
	"strings"
)

const (
	ExitSuccess    = 0
	ExitError      = 1
	ExitAuthError  = 2
	ExitNotFound   = 3
	ExitValidation = 4
)

type APIError struct {
	StatusCode int
	Errors     []APIErrorDetail
}

type APIErrorDetail struct {
	Context       string `json:"context"`
	Message       string `json:"message"`
	ExceptionName string `json:"exceptionName"`
}

func (e *APIError) Error() string {
	if len(e.Errors) > 0 {
		msgs := make([]string, len(e.Errors))
		for i, err := range e.Errors {
			msgs[i] = err.Message
		}
		return fmt.Sprintf("Bitbucket API error (%d): %s", e.StatusCode, strings.Join(msgs, "; "))
	}
	return fmt.Sprintf("Bitbucket API error (%d)", e.StatusCode)
}

func (e *APIError) ExitCode() int {
	switch {
	case e.StatusCode == 401 || e.StatusCode == 403:
		return ExitAuthError
	case e.StatusCode == 404:
		return ExitNotFound
	default:
		return ExitError
	}
}
