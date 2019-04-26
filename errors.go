package errors

import "encoding/json"

const (
	message = "message"
	fields  = "fields"
)

// Error contains current and previous possible errors.
type Error struct {
	msg    string
	fields Fields
	prev   error
}

func (err Error) Error() string {
	return err.msg
}

// Fields is the alias for map[string]interface{}.
type Fields map[string]interface{}

// New creates new error.
func New(msg string) error {
	return &Error{msg: msg}
}

// NewWithFields creates new error with contextual fields.
func NewWithFields(msg string, fields Fields) error {
	return &Error{msg: msg, fields: fields}
}

// Wrap wraps previous error into a new one with explanatory message.
func Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}

	return &Error{msg: msg, prev: err}
}

// WrapWithFields wraps previous error into a new one with explanatory message and with contextual fields.
func WrapWithFields(err error, msg string, fields Fields) error {
	if err == nil {
		return nil
	}

	return &Error{msg: msg, fields: fields, prev: err}
}

// Stack returns a stack of errors in sequential order.
func Stack(err error) []Fields {
	return stack(err, make([]Fields, 0))
}

// JsonStack returns a json-encoded stack of errors.
func JsonStack(err error) []byte {
	b, _ := json.Marshal(Stack(err))
	return b
}

func stack(err error, s []Fields) []Fields {
	if err == nil {
		return s
	}

	item := make(Fields)

	switch e := err.(type) {
	case *Error:
		item[message] = e.msg

		if e.fields != nil && len(e.fields) > 0 {
			item[fields] = e.fields
		}

		s = append(s, item)

		if e.prev != nil {
			s = stack(e.prev, s)
		}
	default:
		item[message] = e.Error()
		s = append(s, item)
	}

	return s
}
