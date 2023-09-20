package errorss

import (
	"errors"
	"fmt"
)

type Error struct {
	code int
	msg  string
	err  error
}

func New(code int, msg string, err error) *Error {
	return &Error{
		code: code,
		msg:  msg,
		err:  err,
	}
}

func (e *Error) Error() string {
	if e.err != nil {
		return fmt.Sprintf("error: error_no = %d error_msg = %s err: %s", e.Code(), e.Msg(), e.err.Error())
	} else {
		return fmt.Sprintf("error: error_no = %d error_msg = %s", e.Code(), e.Msg())
	}
}

func (e *Error) Code() int {
	return e.code
}

func (e *Error) Msg() string {
	return e.msg
}

func (e *Error) Err() error {
	return e.err
}

func FromError(err error) *Error {
	if err == nil {
		return nil
	}

	if se := new(Error); errors.As(err, &se) {
		return se
	}

	return New(SYSTEM_ERROR, ERR_MSG_MAP[SYSTEM_ERROR], err)
}

func NewWithMsg(msg string) *Error {
	return New(SYSTEM_ERROR, msg, nil)
}

func NewWithCode(code int) *Error {
	return New(code, ERR_MSG_MAP[code], nil)
}

func Wrap(err error, msg string) *Error {
	return New(SYSTEM_ERROR, msg, err)
}

func NewWithError(code int, err error) *Error {
	return New(code, ERR_MSG_MAP[code], err)
}

func NewWithCodeMsg(code int, msg string) *Error {
	return New(code, msg, nil)
}
