// Package errors provides functions and types for effective work with errors,
// the wraps, and stack traces.
package errors

import (
	"encoding/json"
	builtin "errors"
	"strings"
)

const (
	message = "message"
	fields  = "fields"
)

// Fields represents just an alias for a more complicated on reading
// map[string]interface{}.
type Fields map[string]interface{}

// Error contains current and previous possible errors.
type Error struct {
	msg    string
	fields map[string]interface{}
	prev   error
}

func (err Error) Error() string {
	stack := messagesStack(&err, make([]string, 0))
	return strings.Join(stack, ": ")
}

// Unwrap returns previous error.
func (err *Error) Unwrap() error {
	return err.prev
}

// Is reports whether any error in err's chain matches target.
func Is(err, target error) bool {
	return builtin.Is(err, target)
}

// As finds the first error in err's chain that matches target, and if so, sets
// target to that error value and returns true.
func As(err error, target interface{}) bool {
	return builtin.As(err, target)
}

func messagesStack(e error, s []string) []string {
	if err, ok := e.(*Error); ok {
		s = append(s, err.msg)
		if err.prev != nil {
			s = messagesStack(err.prev, s)
		}
	} else {
		s = append(s, e.Error())
	}

	return s
}

// New creates new error.
func New(msg string) error {
	return &Error{msg: msg}
}

// NewWithFields creates new error with contextual fields.
func NewWithFields(msg string, fields map[string]interface{}) error {
	return &Error{msg: msg, fields: fields}
}

// Wrap wraps previous error into a new one with explanatory message.
func Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}

	return &Error{msg: msg, prev: err}
}

// WrapWithFields wraps previous error into a new one with explanatory message
// and with contextual fields.
func WrapWithFields(err error, msg string, fields map[string]interface{}) error {
	if err == nil {
		return nil
	}

	return &Error{msg: msg, fields: fields, prev: err}
}

// Stack returns a stack of errors in sequential order.
func Stack(err error) []map[string]interface{} {
	return stack(err, make([]map[string]interface{}, 0))
}

// JSONStack returns a json-encoded stack of errors.
func JSONStack(err error) json.RawMessage {
	b, _ := json.Marshal(Stack(err))
	return b
}

// Restore restores an error from a result returned by Stack func.
func Restore(stack []map[string]interface{}) error {

	var err error

	for i := len(stack) - 1; i >= 0; i-- {

		e := restoreErrorFromStackItem(stack[i])

		if err != nil {
			e.prev = err
		}

		err = e
	}

	return err
}

// Unwrap returns previous error if err's type contains an Unwrap method
// returning error, otherwise, Unwrap returns nil.
// (Doc copied from builtin function.)
func Unwrap(err error) error {
	return builtin.Unwrap(err)
}

func restoreErrorFromStackItem(stackItem map[string]interface{}) *Error {
	err := &Error{}

	if msg, ok := stackItem[message]; ok {
		if s, ok := msg.(string); ok {
			err.msg = s
		}
	}

	if fields, ok := stackItem[fields]; ok {
		if f, ok := fields.(map[string]interface{}); ok {
			err.fields = f
		}
	}

	return err
}

// RestoreRaw restores an error from a result returned by JSONStack func.
func RestoreRaw(stack []byte) error {
	s := make([]map[string]interface{}, 0)
	_ = json.Unmarshal(stack, &s)
	return Restore(s)
}

func stack(err error, s []map[string]interface{}) []map[string]interface{} {
	if err == nil {
		return s
	}

	var item map[string]interface{}

	switch e := err.(type) {
	case *Error:
		item = createStackItemFromError(e)
		s = append(s, item)

		s = stack(e.prev, s)
	default:
		item = make(map[string]interface{})
		item[message] = e.Error()
		s = append(s, item)

		s = stack(builtin.Unwrap(e), s)
	}

	return s
}

func createStackItemFromError(err *Error) map[string]interface{} {
	item := make(map[string]interface{})
	item[message] = err.msg

	if err.fields != nil && len(err.fields) > 0 {
		item[fields] = err.fields
	}

	return item
}
