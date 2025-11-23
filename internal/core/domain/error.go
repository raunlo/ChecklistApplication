package domain

import "fmt"

type Error interface {
	Error() string
	ResponseCode() int
}

type apiError struct {
	errorMessage string
	responseCode int
	parentError  error
}

func (e *apiError) Error() string {
	if e.parentError != nil {
		return fmt.Errorf("%s(%s)", e.errorMessage, e.parentError).Error()
	}
	return e.errorMessage
}

func (e *apiError) ResponseCode() int {
	return e.responseCode
}

func NewError(errorMessage string, responseCode int) Error {
	return &apiError{
		errorMessage: errorMessage,
		responseCode: responseCode,
	}
}

func Wrap(e error, message string, responseCode int) Error {
	return &apiError{
		errorMessage: message,
		responseCode: responseCode,
		parentError:  e,
	}
}
