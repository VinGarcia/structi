package slicei_test

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	tt "github.com/vingarcia/structi/internal/testtools"
	"github.com/vingarcia/structi/slicei"
)

var typeOfString = reflect.TypeOf("")

func TestSliceAppend(t *testing.T) {
	t.Run("should append to a slice correctly", func(t *testing.T) {
		input := []string{"f1", "f2"}
		err := slicei.Append(&input, "f3", "f4")
		tt.AssertNoErr(t, err)

		tt.AssertEqual(t, input, []string{"f1", "f2", "f3", "f4"})
	})

	t.Run("should convert values appropriately", func(t *testing.T) {
		input := []uint{1, 2}
		err := slicei.Append(&input, int(3), int8(4))
		tt.AssertNoErr(t, err)

		tt.AssertEqual(t, input, []uint{1, 2, 3, 4})
	})

	t.Run("should do nothing with empty inputs", func(t *testing.T) {
		input := []int{1, 2, 3}
		err := slicei.Append(&input)
		tt.AssertNoErr(t, err)

		tt.AssertEqual(t, input, []int{1, 2, 3})
	})

	t.Run("should work with nil slices", func(t *testing.T) {
		var input []int
		err := slicei.Append(&input, 42)
		tt.AssertNoErr(t, err)

		tt.AssertEqual(t, input, []int{42})
	})

	t.Run("should report error if input is not a type of slice", func(t *testing.T) {
		input := map[string]uint{}
		err := slicei.Append(&input, "foo")
		tt.AssertErrContains(t, err, "slice", "map[string]uint")
	})

	t.Run("should report error if conversion of one of the items is not possible", func(t *testing.T) {
		input := []int{}
		err := slicei.Append(&input, struct{}{})
		tt.AssertErrContains(t, err, "converting", "int", "struct")
	})

	t.Run("should report error if input slice is nil", func(t *testing.T) {
		err := slicei.Append(nil, 42)
		tt.AssertErrContains(t, err, "unexpected nil input")
	})
}

func TestSliceForEach(t *testing.T) {
	t.Run("should iterate over a simple array correctly", func(t *testing.T) {
		input := []string{"s1", "s2", "s3"}

		output := []slicei.Field{}
		err := slicei.ForEach(&input, func(f slicei.Field) error {
			output = append(output, f)
			return nil
		})
		tt.AssertNoErr(t, err)

		for i := range output {
			tt.AssertEqual(t, output[i].Set == nil, false)
			output[i].Set = nil
		}

		tt.AssertEqual(t, output, []slicei.Field{
			{
				Index: 0,
				Kind:  reflect.String,
				Type:  typeOfString,
				Value: ptr("s1"),
			},
			{
				Index: 1,
				Kind:  reflect.String,
				Type:  typeOfString,
				Value: ptr("s2"),
			},
			{
				Index: 2,
				Kind:  reflect.String,
				Type:  typeOfString,
				Value: ptr("s3"),
			},
		})
	})

	t.Run("should allow setting values on the input slice", func(t *testing.T) {
		input := []string{"s1", "s2", "s3"}

		i := 0
		err := slicei.ForEach(&input, func(f slicei.Field) error {
			i++
			return f.Set("new" + strconv.Itoa(i))
		})
		tt.AssertNoErr(t, err)

		tt.AssertEqual(t, input, []string{"new1", "new2", "new3"})
	})

	t.Run("should not panic nor do anything for nil slices", func(t *testing.T) {
		var input []string

		i := 0
		err := slicei.ForEach(&input, func(f slicei.Field) error {
			i++
			return fmt.Errorf("should not run")
		})
		tt.AssertNoErr(t, err)
		tt.AssertEqual(t, i, 0)
	})

	t.Run("validation errors", func(t *testing.T) {
		tests := []struct {
			desc               string
			input              any
			fn                 func(item slicei.Field) error
			expectErrToContain []string
		}{
			{
				desc:               "should report error if input is not a pointer",
				input:              []string{"s1"},
				expectErrToContain: []string{"expected slice pointer", "[]string"},
			},
			{
				desc:               "should report error if input is not a slice",
				input:              ptr(map[string]any{"k1": "v1"}),
				expectErrToContain: []string{"slice", "map[string]interface {}"},
			},
			{
				desc:               "should report error if input is nil",
				input:              nil,
				expectErrToContain: []string{"unexpected nil input"},
			},
			{
				desc:  "should report error if Set() is called with an invalid value",
				input: ptr([]string{"s1"}),
				fn: func(item slicei.Field) error {
					return item.Set(struct{}{})
				},
				expectErrToContain: []string{"[]string[0]", "struct", "cannot convert"},
			},
		}

		for _, test := range tests {
			t.Run(test.desc, func(t *testing.T) {
				err := slicei.ForEach(test.input, func(f slicei.Field) error {
					if test.fn != nil {
						return test.fn(f)
					}
					return nil
				})
				tt.AssertErrContains(t, err, test.expectErrToContain...)
			})
		}
	})
}

func ptr[T any](t T) *T {
	return &t
}
