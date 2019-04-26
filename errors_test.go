package errors_test

import (
	"bytes"
	"encoding/json"
	builtin "errors"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/krocos/errors"
)

func TestJsonStack(t *testing.T) {
	readJson := createJsonReader(t, "testdata")

	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want []byte
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
			want: readJson("Stack_OnlyPackageError_Simple.json"),
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
			want: readJson("Stack_OnlyPackageError_StartWithFields.json"),
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
			want: readJson("Stack_BuiltinErrors.json"),
		},
		{
			name: "Stack_NilError",
			args: args{
				err: func() error {
					var err error
					err = nil
					err = errors.Wrap(err, "2")
					err = errors.WrapWithFields(err, "3", errors.Fields{
						"f1": "v1",
					})

					return err
				}(),
			},
			want: readJson("Stack_NilError.json"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := errors.JsonStack(tt.args.err); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("JsonStack() = %v, want %v", got, tt.want)
			}
		})
	}
}

func createJsonReader(t *testing.T, dirname string) func(filename string) []byte {
	t.Helper()

	return func(filename string) []byte {
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
