package errors

import "encoding/json"

const (
	message = "message"
	fields  = "fields"
)

// Fields represents just an alias for a more complicated on reading map[string]interface{}.
type Fields map[string]interface{}

// Error contains current and previous possible errors.
type Error struct {
	msg    string
	fields map[string]interface{}
	prev   error
}

func (err Error) Error() string {
	return err.msg
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

// WrapWithFields wraps previous error into a new one with explanatory message and with contextual fields.
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

// JsonStack returns a json-encoded stack of errors.
func JsonStack(err error) []byte {
	b, _ := json.Marshal(Stack(err))
	return b
}

// Restore restores an error from a result returned by Stack func.
func Restore(stack []map[string]interface{}) error {

	var err error

	for i := len(stack) - 1; i >= 0; i-- {

		e := &Error{}

		if msg, ok := stack[i][message]; ok {
			if s, ok := msg.(string); ok {
				e.msg = s
			}
		}

		if fields, ok := stack[i][fields]; ok {
			if f, ok := fields.(map[string]interface{}); ok {
				e.fields = f
			}
		}

		if err != nil {
			e.prev = err
		}

		err = e
	}

	return err
}

// RestoreRaw restores an error from a result returned by JsonStack func.
func RestoreRaw(stack []byte) error {
	s := make([]map[string]interface{}, 0)
	_ = json.Unmarshal(stack, &s)
	return Restore(s)
}

func stack(err error, s []map[string]interface{}) []map[string]interface{} {
	if err == nil {
		return s
	}

	item := make(map[string]interface{})

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
