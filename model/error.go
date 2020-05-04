package model

import (
	"fmt"
	"log"
	"strings"

	"github.com/go-playground/validator/v10"
)

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

// Errors is an ordered collection of errors
type Errors struct {
	errs       []Error
	httpStatus int
}

// NewErrors buids an Errors collection from an error, which may actually be a ValidationErrors colletion
func NewErrors(httpStatus int, err error) *Errors {
	errors := Errors{
		errs:       make([]Error, 0),
		httpStatus: httpStatus,
	}
	if ves, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range ves {
			if fe.Tag() == "required" {
				name := strings.SplitN(fe.Namespace(), ".", 2)
				// log.Printf("name: %v", name)
				errors.errs = append(errors.errs, NewError(ErrRequired, name[1]))
			} else {
				errors.errs = append(errors.errs, NewError(ErrOther, fmt.Sprintf("Key: '%s' Error:Field validation for '%s' failed on the '%s' tag", fe.Namespace(), fe.Field(), fe.Tag())))
			}
		}
	} else {
		errors.errs = append(errors.errs, NewError(ErrOther, err.Error()))
	}
	return &errors
}

// HTTPStatus returns the HTTP status code
func (e Errors) HTTPStatus() int {
	return e.httpStatus
}

// Errs returns the slice of Error structs
func (e Errors) Errs() []Error {
	return e.errs
}

// func (e Errors) Error() string {
// 	msg := "Errors:"
// 	for _, er := range e.errs {
// 		params := make([]interface{}, len(er.Params))
// 		for i, p := range er.Params {
// 			params[i] = p
// 		}
// 		msg += "\n  " + fmt.Sprintf(er.Message, params...)
// 	}
// 	return msg
// }
