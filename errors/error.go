package errors

import (
	"fmt"
	"strings"
)

type Error struct {
	// code is the error code of the error. When marshaled to JSON, it will be a string.
	code Code

	// msg is the user-friendly message returned to the client.
	msg string

	// op operation where error occured
	op string

	// details is the internal error message returned to the developer.
	details []any

	// underlying error
	cause *Error

	// depth of the error tree
	depth int
}

func New(msg string) *Error {
	return &Error{msg: msg}
}

func Newf(msg string, args ...interface{}) *Error {
	return &Error{msg: fmt.Sprintf(msg, args...)}
}

func (err *Error) Code(code Code) *Error {
	err.code = code
	return err
}

func (err *Error) Op(op string) *Error {
	err.op = op
	return err
}

func (err *Error) Details(details ...any) *Error {
	err.details = details
	return err
}

// Error returns the error in the format "code: message".
func (e *Error) Error() string {
	if e.op == "" {
		return fmt.Sprintf("%s: %s", e.code.String(), e.msg)
	}
	return fmt.Sprintf("%s: %s: %s", e.code.String(), e.op, e.msg)
}

// Stack returns a description of the error and all it's underlying errors.
func (e *Error) Stack() []string {
	stack := make([]string, e.depth+1)

	for i, err := 0, e; err != nil; err, i = err.cause, i+1 {
		tabOffset := strings.Repeat("\t", i)

		var buf strings.Builder
		write := func(s string) {
			buf.WriteString(tabOffset)
			buf.WriteString(s + "\n")
		}

		write(err.Error())
		for j, d := range err.details {
			write(fmt.Sprintf("\t%d: %+v", j, d))
		}

		stack[i] = buf.String()
	}
	return stack
}

// Unwrap returns the underlying error.
func (e *Error) Unwrap() error {
	if e.cause == nil {
		return nil
	}
	return e.cause
}

func (e *Error) Is(target error) bool {
	if t, ok := target.(*Error); ok {
		return equalNodes(e, t)
	}
	return false
}

// Wrap wraps an underlying error `child` with a new error `parent`.
//
// - when the child error is nil, the parent error is returned as is.
//
// - when the parent error is nil, the child error is returned as is.
//
// - when both errors are nil, nil is returned.
func Wrap(child, parent error) error {
	parent = Convert(parent)
	child = Convert(child)
	switch {
	case parent == nil && child == nil:
		return nil
	case parent == nil:
		return child
	case child == nil:
		return parent
	default:
		p := parent.(*Error)
		c := child.(*Error)
		p.cause = c
		p.depth = c.depth + 1
		return p
	}
}

// Convert converts any error to an *Error type. If the error is already an *Error, it is returned as is.
// nil errors are returned as nil.
func Convert(err error) error {
	if err == nil {
		return nil
	}
	if e, ok := err.(*Error); ok {
		return e
	}
	return &Error{
		code: Unknown,
		msg:  err.Error(),
	}
}

// equalNodes was created because we can't even trust go to compare equality of the error structs.
// Comparison does not involve the underlying errors because we don't want to compare the entire error tree.
//
// The fields considered for equality are error codes and messages. It makes sense to leave details out because two errors
// might be the same but with different details.
func equalNodes(a, b *Error) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if a.code != b.code {
		return false
	}
	if a.op != b.op {
		return false
	}
	if len(a.msg) != len(b.msg) {
		return false
	}
	for i := range a.msg {
		if a.msg[i] != b.msg[i] {
			return false
		}
	}
	return true
}
