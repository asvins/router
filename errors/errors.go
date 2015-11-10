package errors

import (
	"net/http"
)

type Http interface {
	Message() string
	Code() int
}

/*
*	HTTP status BadRequest
 */
// BadRequestStruct http error
type BadRequestStruct struct {
	Msg string `json:"message"`
}

// BadRequest returns a newly allocated BadRequestStruct
func BadRequest(msg string) BadRequestStruct {
	return BadRequestStruct{msg}
}

// Message - needed to implement HttpErros interface
func (e BadRequestStruct) Message() string {
	return e.Msg
}

// Code - needed to implement Http interface
func (e BadRequestStruct) Code() int {
	return http.StatusBadRequest
}

/*
*	HTTP status Unauthorized
 */
// UnauthorizedStruct http error
type UnauthorizedStruct struct {
	Msg string `json:"message"`
}

// Unauthorizes returns a newly allocated UnauthorizedStruct
func Unauthorized(msg string) UnauthorizedStruct {
	return UnauthorizedStruct{msg}
}

// Message - needed to implement HttpErros interface
func (e UnauthorizedStruct) Message() string {
	return e.Msg
}

// Code - needed to implement Http interface
func (e UnauthorizedStruct) Code() int {
	return http.StatusUnauthorized
}

/*
*	HTTP status NotFound
 */
// NotFoundStruct http error
type NotFoundStruct struct {
	Msg string `json:"message"`
}

// NotFound returns a newly allocated NotFoundStruct
func NotFound(message string) NotFoundStruct {
	return NotFoundStruct{message}
}

// Message - needed to implement HttpErros interface
func (e NotFoundStruct) Message() string {
	return e.Msg
}

// Code - needed to implement Http interface
func (e NotFoundStruct) Code() int {
	return http.StatusNotFound
}
