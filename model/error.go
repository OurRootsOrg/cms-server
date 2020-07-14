package model

import (
	"fmt"
	"log"
)

// ErrorCode is one of the valid error codes the API can return
type ErrorCode string

// Matches returns `true` if the passed `error` is a `model.Error` with a matching `ErrorCode`
func (code ErrorCode) Matches(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Code == code
	}
	return false
}

// Standard error codes
const (
	ErrRequired         ErrorCode = "REQUIRED"
	ErrNotFound         ErrorCode = "NOT_FOUND"
	ErrBadReference     ErrorCode = "BAD_REFERENCE"
	ErrConcurrentUpdate ErrorCode = "CONCURRENT_UPDATE"
	ErrOther            ErrorCode = "OTHER"
)

var errorMessages = map[ErrorCode]string{
	ErrRequired:         "Field '%s' is required",
	ErrNotFound:         "'%s' was not found",
	ErrBadReference:     "Non-existent reference. ID: '%s', Type: '%s'",
	ErrConcurrentUpdate: "Database LastUpdateTime (%s) doesn't match provided value (%s).",
	ErrOther:            "Unknown error: %s",
}

// Error represents a single API error
type Error struct {
	Code    ErrorCode `json:"code"`
	Params  []string  `json:"params"`
	Message string    `json:"message"`
}

// NewError build an error. If the error code is unknown it is set to ErrOther.
func NewError(code ErrorCode, params ...string) *Error {
	msg, ok := errorMessages[code]
	if !ok {
		log.Printf("[INFO] Unknown error code '%s', setting to ErrOther", code)
		code = ErrOther
		msg = errorMessages[code]
	}
	return &Error{
		Code:    code,
		Message: msg,
		Params:  params,
	}
}

func (e Error) Error() string {
	params := make([]interface{}, len(e.Params))
	for i, p := range e.Params {
		params[i] = p
	}
	return fmt.Sprintf("Error %s: ", e.Code) + fmt.Sprintf(e.Message, params...)
}
