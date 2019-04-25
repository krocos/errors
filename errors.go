package errors

import "encoding/json"

const (
	message = "message"
	fields  = "fields"
)

type Error struct {
	msg    string
	fields Fields
	prev   error
}

type Fields map[string]interface{}

func (err Error) Error() string {
	return err.msg
}

func New(msg string) error {
	return &Error{msg: msg}
}

func NewWithFields(msg string, fields Fields) error {
	return &Error{msg: msg, fields: fields}
}

func Wrap(err error, msg string) error {
	return &Error{msg: msg, prev: err}
}

func WrapWithFields(err error, msg string, fields Fields) error {
	return &Error{msg: msg, fields: fields, prev: err}
}

func Stack(err error) []Fields {
	return stack(err, make([]Fields, 0))
}

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
