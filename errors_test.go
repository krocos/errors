package errors_test

import (
	"bytes"
	"encoding/json"
	builtin "errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/krocos/errors"
)

func ExampleRestoreRaw() {
	err := errors.NewWithFields("1", errors.Fields{
		"f1": "v1",
	})
	err = errors.WrapWithFields(err, "2", errors.Fields{
		"f2": "v2",
	})

	b := errors.JSONStack(err)
	fmt.Println(string(b))

	// Restore error from a []byte (json) stack.
	err = errors.RestoreRaw(b)
	if err != nil {
		fmt.Println(err.Error())
	}

	// output:
	// [{"fields":{"f2":"v2"},"message":"2"},{"fields":{"f1":"v1"},"message":"1"}]
	// 2: 1
}

func BenchmarkStack_Builtin(b *testing.B) {
	// 5000000 ops
	// 258 ns/op

	err := builtin.New("error")

	for i := 0; i < b.N; i++ {
		_ = errors.Stack(err)
	}
}

func BenchmarkJsonStack_Builtin(b *testing.B) {
	// 1000000 ops
	// 1017 ns/op

	err := builtin.New("error")

	for i := 0; i < b.N; i++ {
		_ = errors.JSONStack(err)
	}
}

func TestError_Error_JoinsMessagesCorrectly(t *testing.T) {
	fields := errors.Fields{
		"f1": "v1",
		"f2": "v2",
	}

	err := errors.NewWithFields("first message", fields)
	err = errors.Wrap(err, "second msg")
	err = errors.WrapWithFields(err, "third message", fields)

	if err == nil {
		t.Fatalf("err must not be nil")
	}

	want := "third message: second msg: first message"
	got := err.Error()

	if want != got {
		t.Fatalf("want the string `%s`, got `%s`", want, got)
	}
}

func TestJsonStack(t *testing.T) {
	readJSON := createJSONReader(t, "testdata")

	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want json.RawMessage
	}{
		{
			name: "Stack_OnlyPackageError_Simple",
			args: args{
				err: func() error {
					err := errors.New("1")
					err = errors.Wrap(err, "2")
					err = errors.WrapWithFields(err, "3", errors.Fields{
						"f1": "v1",
					})

					return err
				}(),
			},
			want: readJSON("Stack_OnlyPackageError_Simple.json"),
		},
		{
			name: "Stack_OnlyPackageError_StartWithFields",
			args: args{
				err: func() error {
					err := errors.NewWithFields("1", errors.Fields{
						"f1": "v1",
					})
					err = errors.WrapWithFields(err, "2", errors.Fields{
						"f1": "v1",
					})
					err = errors.Wrap(err, "3")

					return err
				}(),
			},
			want: readJSON("Stack_OnlyPackageError_StartWithFields.json"),
		},
		{
			name: "Stack_BuiltinErrors",
			args: args{
				err: func() error {
					err := builtin.New("1")
					err = errors.Wrap(err, "2")
					err = errors.WrapWithFields(err, "3", errors.Fields{
						"f1": "v1",
					})

					return err
				}(),
			},
			want: readJSON("Stack_BuiltinErrors.json"),
		},
		{
			name: "Stack_NilError",
			args: args{
				err: func() error {
					var err error
					err = errors.Wrap(err, "2")
					err = errors.WrapWithFields(err, "3", errors.Fields{
						"f1": "v1",
					})

					return err
				}(),
			},
			want: readJSON("Stack_NilError.json"),
		},
	}
	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			if got := errors.JSONStack(tc.args.err); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("JSONStack() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestRestore(t *testing.T) {

	const (
		emptyStack                  = "RestoreEmptyStack"
		fullStack                   = "RestoreFullStack"
		fullStackStartedFromBuiltin = "RestoreFullStackStartedFromBuiltin"
	)

	// []map[string]interface{} is a stack.
	stacks := map[string][]map[string]interface{}{
		emptyStack: func() []map[string]interface{} {
			return make([]map[string]interface{}, 0)
		}(),
		fullStack: func() []map[string]interface{} {
			err := errors.NewWithFields("1", errors.Fields{
				"f1": "v1",
			})
			err = errors.WrapWithFields(err, "2", errors.Fields{
				"f1": "v1",
			})

			return errors.Stack(err)
		}(),
		fullStackStartedFromBuiltin: func() []map[string]interface{} {
			err := builtin.New("1")
			err = errors.WrapWithFields(err, "2", errors.Fields{
				"f1": "v1",
			})

			return errors.Stack(err)
		}(),
	}

	type args struct {
		stack []map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want []map[string]interface{}
	}{
		{
			name: emptyStack,
			args: args{
				stack: stacks[emptyStack],
			},
			want: stacks[emptyStack],
		},
		{
			name: fullStack,
			args: args{
				stack: stacks[fullStack],
			},
			want: stacks[fullStack],
		},
		{
			name: fullStackStartedFromBuiltin,
			args: args{
				stack: stacks[fullStackStartedFromBuiltin],
			},
			want: stacks[fullStackStartedFromBuiltin],
		},
	}
	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			err := errors.Restore(tc.args.stack)
			got := errors.Stack(err)

			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Restore() = %v, want: %v", got, tc.want)
			}
		})
	}
}

func TestRestoreRaw(t *testing.T) {

	const (
		emptyStack                  = "RestoreEmptyStack"
		fullStack                   = "RestoreFullStack"
		fullStackStartedFromBuiltin = "RestoreFullStackStartedFromBuiltin"
	)

	// []byte is a raw stack.
	stacks := map[string]json.RawMessage{
		emptyStack: func() json.RawMessage {
			return []byte("[]")
		}(),
		fullStack: func() json.RawMessage {
			err := errors.NewWithFields("1", errors.Fields{
				"f1": "v1",
			})
			err = errors.WrapWithFields(err, "2", errors.Fields{
				"f1": "v1",
			})

			return errors.JSONStack(err)
		}(),
		fullStackStartedFromBuiltin: func() json.RawMessage {
			err := builtin.New("1")
			err = errors.WrapWithFields(err, "2", errors.Fields{
				"f1": "v1",
			})

			return errors.JSONStack(err)
		}(),
	}

	type args struct {
		stack []byte
	}
	tests := []struct {
		name string
		args args
		want json.RawMessage
	}{
		{
			name: emptyStack,
			args: args{
				stack: stacks[emptyStack],
			},
			want: stacks[emptyStack],
		},
		{
			name: fullStack,
			args: args{
				stack: stacks[fullStack],
			},
			want: stacks[fullStack],
		},
		{
			name: fullStackStartedFromBuiltin,
			args: args{
				stack: stacks[fullStackStartedFromBuiltin],
			},
			want: stacks[fullStackStartedFromBuiltin],
		},
	}
	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			err := errors.RestoreRaw(tc.args.stack)
			got := errors.JSONStack(err)

			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("RestoreRaw() = %s, want: %v", string(got), string(tc.want))
			}
		})
	}
}

func createJSONReader(t *testing.T, dirname string) func(filename string) json.RawMessage {
	t.Helper()

	return func(filename string) json.RawMessage {
		t.Helper()

		p := path.Join(dirname, filename)
		file, err := os.Open(p)
		if err != nil {
			t.Fatalf("failed to open the file '%s': %v", p, err)
		}

		b, err := ioutil.ReadAll(file)
		if err != nil {
			t.Fatalf("failed to read the file '%s': %v", p, err)
		}

		buf := &bytes.Buffer{}

		err = json.Compact(buf, b)
		if err != nil {
			t.Fatalf("failed to compact json: %v", err)
		}

		return buf.Bytes()
	}
}
