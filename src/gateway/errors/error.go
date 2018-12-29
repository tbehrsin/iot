package errors

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type Error struct {
	code   int
	source error
}

func NewError(code int, err ...interface{}) Error {
	if len(err) == 0 {
		return Error{code, nil}
	}

	if e, ok := err[0].(Error); ok {
		return e
	}

	if e, ok := err[0].(error); ok {
		return Error{code, e}
	}

	if err[0] == nil {
		return Error{http.StatusInternalServerError, nil}
	}

	if f, ok := err[0].(string); ok {
		return Error{code, fmt.Errorf(f, err[1:]...)}
	}

	panic(fmt.Errorf("unable to format error arguments: %+v", err))
}

func NewUnauthorized(err ...interface{}) Error {
	return NewError(http.StatusUnauthorized, err...)
}

func NewForbidden(err ...interface{}) Error {
	return NewError(http.StatusForbidden, err...)
}

func NewNotFound(err ...interface{}) Error {
	return NewError(http.StatusNotFound, err...)
}

func NewBadRequest(err ...interface{}) Error {
	return NewError(http.StatusBadRequest, err...)
}

func NewInternalServerError(err ...interface{}) Error {
	return NewError(http.StatusInternalServerError, err...)
}

func (e Error) Error() string {
	return fmt.Sprintf("%+v", e.source)
}

func (e Error) Write(w io.Writer) {
	if rw, ok := w.(http.ResponseWriter); ok {
		rw.WriteHeader(e.code)
	}

	buf, _ := e.MarshalJSON()
	w.Write(buf)
}

func (e Error) Println() Error {
	log.Printf("%+v\n", e)
	return e
}

func (e Error) MarshalJSON() ([]byte, error) {
	out := make(map[string]interface{})
	error := make(map[string]interface{})
	out["error"] = error

	error["message"] = strings.ToLower(http.StatusText(e.code))
	error["code"] = e.code

	return json.Marshal(out)
}
