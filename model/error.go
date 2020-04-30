package model

import "log"

// ErrorCode is one of the valid error codes the API can return
type ErrorCode string

// Standard error codes
const (
	ErrRequired ErrorCode = "REQUIRED"
	ErrOther    ErrorCode = "OTHER"
)

var errorMessages = map[ErrorCode]string{
	ErrRequired: "Field '%s' is required",
	ErrOther:    "Unknown error: %s",
}

// Errors is an ordered collection of errors
type Errors []Error

// Error represents a single API error
type Error struct {
	Code    ErrorCode `json:"code"`
	Params  []string  `json:"params"`
	Message string    `json:"message"`
}

// NewError build an error. If the error code is unknown it is set to ErrOther.
func NewError(code ErrorCode, params ...string) Error {
	msg, ok := errorMessages[code]
	if !ok {
		log.Printf("[INFO] Unknown error code '%s', setting to ErrOther", code)
		code = ErrOther
		msg = errorMessages[code]
	}
	return Error{
		Code:    code,
		Message: msg,
		Params:  params,
	}
}
