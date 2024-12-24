package structi_test

import (
	"errors"
	"reflect"
	"strconv"
	"testing"

	"github.com/vingarcia/structi"
	tt "github.com/vingarcia/structi/internal/testtools"
)

func TestForEach(t *testing.T) {
	t.Run("should parse a single tag with a hardcoded value", func(t *testing.T) {
		var output struct {
			Attr1 string `env:"attr1"`
		}
		err := structi.ForEach(&output, func(f structi.Field) error {
			return f.Set("fake-value-for-string")
		})
		tt.AssertNoErr(t, err)
		tt.AssertEqual(t, output.Attr1, "fake-value-for-string")
	})

	t.Run("should ignore attributes if the function returns a nil value", func(t *testing.T) {
		var output struct {
			Attr1 string `env:"attr1"`
			Attr2 string `someothertag:"attr2"`
		}
		output.Attr2 = "placeholder"
		err := structi.ForEach(&output, func(field structi.Field) error {
			envTag := field.Tags["env"]
			if envTag == "" {
				return nil
			}

			return field.Set("fake-value-for-string")
		})
		tt.AssertNoErr(t, err)
		tt.AssertEqual(t, output.Attr1, "fake-value-for-string")
		tt.AssertEqual(t, output.Attr2, "placeholder")
	})

	t.Run("should be able to fill multiple attributes", func(t *testing.T) {
		var output struct {
			Attr1 string `map:"f1"`
			Attr2 string `map:"f2"`
			Attr3 string `map:"f3"`
		}
		err := structi.ForEach(&output, func(field structi.Field) error {
			v := map[string]string{
				"f1": "v1",
				"f2": "v2",
				"f3": "v3",
			}[field.Tags["map"]]

			return field.Set(v)
		})
		tt.AssertNoErr(t, err)
		tt.AssertEqual(t, output.Attr1, "v1")
		tt.AssertEqual(t, output.Attr2, "v2")
		tt.AssertEqual(t, output.Attr3, "v3")
	})

	t.Run("should ignore private fields", func(t *testing.T) {
		var output struct {
			Attr1 string `env:"attr1"`
			attr2 string `env:"attr2"`
		}
		err := structi.ForEach(&output, func(field structi.Field) error {
			return field.Set("fake-value-for-string")
		})
		tt.AssertNoErr(t, err)
		tt.AssertEqual(t, output.Attr1, "fake-value-for-string")
		tt.AssertEqual(t, output.attr2, "")
	})

	t.Run("nested structs", func(t *testing.T) {
		t.Run("should parse fields recursively", func(t *testing.T) {
			var output struct {
				Attr1       int `env:"attr1"`
				OtherStruct struct {
					Attr2 int `env:"attr1"`
				}
			}
			err := structi.ForEach(&output, func(field structi.Field) error {
				if field.Kind == reflect.Struct {
					return structi.ForEach(field.Value, func(field structi.Field) error {
						return field.Set(42)
					})
				}

				return field.Set(64)
			})
			tt.AssertNoErr(t, err)
			tt.AssertEqual(t, output.Attr1, 64)
			tt.AssertEqual(t, output.OtherStruct.Attr2, 42)
		})

		/*
			t.Run("should parse fields recursively even for nil pointers to struct", func(t *testing.T) {
				decoder := structi.FuncTagDecoder(func(field structi.Field) (interface{}, error) {
					if field.Kind == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct {
						return structi.FuncTagDecoder(func(field structi.Field) (interface{}, error) {
							return 42, nil
						}), nil
					}

					return 64, nil
				})

				var output struct {
					Attr1       int `env:"attr1"`
					OtherStruct *struct {
						Attr2 int `env:"attr2"`
					}
				}
				err := structi.ForEach(&output, decoder)
				tt.AssertNoErr(t, err)
				tt.AssertEqual(t, output.Attr1, 64)
				tt.AssertEqual(t, output.OtherStruct.Attr2, 42)
			})
		*/

		t.Run("should report error correctly for invalid nested values", func(t *testing.T) {
			tests := []struct {
				desc               string
				targetStruct       interface{}
				expectErrToContain []string
			}{
				{
					desc: "not a struct",
					targetStruct: &struct {
						NotAStruct int `env:"attr1"`
					}{},
					expectErrToContain: []string{"NotAStruct", "can only get struct info from structs", "int"},
				},
				{
					desc: "pointer to not a struct",
					targetStruct: &struct {
						NotAStruct *int `env:"attr1"`
					}{},
					expectErrToContain: []string{"NotAStruct", "can only get struct info from structs", "*int"},
				},
			}
			for _, test := range tests {
				t.Run(test.desc, func(t *testing.T) {
					err := structi.ForEach(test.targetStruct, func(field structi.Field) error {
						// Some tag decoder:
						return structi.ForEach(field.Value, func(field structi.Field) error {
							return field.Set(42)
						})
					})
					tt.AssertErrContains(t, err, test.expectErrToContain...)
				})
			}
		})
	})

	t.Run("nested slices", func(t *testing.T) {
		t.Run("should convert each item of a slice", func(t *testing.T) {
			var output struct {
				Slice []int `map:"slice"`
			}
			err := structi.ForEach(&output, func(field structi.Field) error {
				return field.Set([]interface{}{1, 2, 3})
			})
			tt.AssertNoErr(t, err)
			tt.AssertEqual(t, output.Slice, []int{1, 2, 3})
		})

		t.Run("should convert each item of a slice even with different types", func(t *testing.T) {
			var output struct {
				Slice []float64 `map:"slice"`
			}
			err := structi.ForEach(&output, func(field structi.Field) error {
				return field.Set([]interface{}{1, 2, 3})
			})
			tt.AssertNoErr(t, err)
			tt.AssertEqual(t, output.Slice, []float64{1.0, 2.0, 3.0})
		})

		t.Run("should work with slices of pointers", func(t *testing.T) {
			var output struct {
				Slice []int `map:"slice"`
			}
			err := structi.ForEach(&output, func(field structi.Field) error {
				return field.Set([]*int{
					intPtr(1),
					intPtr(2),
					intPtr(3),
				})
			})
			tt.AssertNoErr(t, err)
			tt.AssertEqual(t, output.Slice, []int{1, 2, 3})
		})

		t.Run("should work with slices of pointers or different types", func(t *testing.T) {
			var output struct {
				Slice []float64 `map:"slice"`
			}
			err := structi.ForEach(&output, func(field structi.Field) error {
				return field.Set([]*int{
					intPtr(1),
					intPtr(2),
					intPtr(3),
				})
			})
			tt.AssertNoErr(t, err)
			tt.AssertEqual(t, output.Slice, []float64{1.0, 2.0, 3.0})
		})

		t.Run("should work with pointers to slices", func(t *testing.T) {
			t.Run("source pointer target non-pointer", func(t *testing.T) {
				var output struct {
					Slice []int `map:"slice"`
				}
				err := structi.ForEach(&output, func(field structi.Field) error {
					return field.Set(&[]int{1, 2, 3})
				})
				tt.AssertNoErr(t, err)
				tt.AssertEqual(t, output.Slice, []int{1, 2, 3})
			})

			t.Run("source non-pointer target pointer", func(t *testing.T) {
				var output struct {
					Slice *[]int `map:"slice"`
				}
				err := structi.ForEach(&output, func(field structi.Field) error {
					return field.Set([]int{1, 2, 3})
				})
				tt.AssertNoErr(t, err)
				tt.AssertEqual(t, output.Slice, &[]int{1, 2, 3})
			})
		})

		t.Run("should work with slices of nested structs/maps", func(t *testing.T) {
			t.Run("input and target being a slice of maps", func(t *testing.T) {
				var output struct {
					Slice []map[string]any `map:"slice"`
				}
				err := structi.ForEach(&output, func(field structi.Field) error {
					return field.Set([]map[string]any{
						{
							"name": "fakeAttrName",
						},
					})
				})
				tt.AssertNoErr(t, err)
				tt.AssertEqual(t, output.Slice, []map[string]any{
					{
						"name": "fakeAttrName",
					},
				})
			})
		})
	})

	t.Run("should convert types correctly", func(t *testing.T) {
		t.Run("should convert different types of integers", func(t *testing.T) {
			var output struct {
				Attr1 int `env:"attr1"`
			}
			err := structi.ForEach(&output, func(field structi.Field) error {
				return field.Set(uint64(10))
			})
			tt.AssertNoErr(t, err)
			tt.AssertEqual(t, output.Attr1, 10)
		})

		t.Run("should convert from ptr to non ptr", func(t *testing.T) {
			var output struct {
				Attr1 int `env:"attr1"`
			}
			err := structi.ForEach(&output, func(field structi.Field) error {
				i := 64
				return field.Set(&i)
			})
			tt.AssertNoErr(t, err)
			tt.AssertEqual(t, output.Attr1, 64)
		})

		t.Run("should convert from ptr to non ptr", func(t *testing.T) {
			var output struct {
				Attr1 *int `env:"attr1"`
			}
			err := structi.ForEach(&output, func(field structi.Field) error {
				return field.Set(64)
			})
			tt.AssertNoErr(t, err)
			tt.AssertNotEqual(t, output.Attr1, nil)
			tt.AssertEqual(t, *output.Attr1, 64)
		})

		t.Run("should work with structs", func(t *testing.T) {
			type Foo struct {
				Name string
			}

			var output struct {
				Attr1 Foo `env:"attr1"`
			}
			err := structi.ForEach(&output, func(field structi.Field) error {
				return field.Set(Foo{
					Name: "test",
				})
			})
			tt.AssertNoErr(t, err)
			tt.AssertEqual(t, output.Attr1, Foo{
				Name: "test",
			})
		})

		t.Run("should work with embeded fields", func(t *testing.T) {
			type Foo struct {
				Name      string
				IsEmbeded bool
			}

			var output struct {
				Foo `env:"attr1"`
			}
			err := structi.ForEach(&output, func(field structi.Field) error {
				return field.Set(Foo{
					Name:      field.Name,      // should be foo
					IsEmbeded: field.IsEmbeded, // should be true
				})
			})
			tt.AssertNoErr(t, err)
			tt.AssertEqual(t, output.Foo, Foo{
				Name:      "Foo",
				IsEmbeded: true,
			})
		})
	})

	t.Run("should report errors correctly", func(t *testing.T) {
		tests := []struct {
			desc               string
			value              interface{}
			targetStruct       interface{}
			expectErrToContain []string
		}{
			{
				desc:               "should report error input is a ptr to something else than a struct",
				value:              "example-value",
				targetStruct:       &[]int{},
				expectErrToContain: []string{"can only get struct info from structs", "[]int"},
			},
			{
				desc:  "should report error if input is not a pointer",
				value: "example-value",
				targetStruct: struct {
					Attr1 string `some_tag:""`
				}{},
				expectErrToContain: []string{"expected struct pointer"},
			},
			{
				desc:  "should report error if input a nil ptr to struct",
				value: "example-value",
				targetStruct: (*struct {
					Attr1 string `some_tag:""`
				})(nil),
				expectErrToContain: []string{"expected non-nil pointer"},
			},
			{
				desc:  "should report error if the type doesnt match",
				value: "example-value",
				targetStruct: &struct {
					Attr1 int `env:"attr1"`
				}{},
				expectErrToContain: []string{"string", "int"},
			},
			{
				desc:  "should report error if parsing a non-slice into a slice field",
				value: "example-value",
				targetStruct: &struct {
					Attr1 []string `some_tag:"attr1"`
				}{},
				expectErrToContain: []string{"iteration error", "Attr1", "string", "[]string", "example-value"},
			},
			// {
			// desc:  "should report error if the conversion fails for one of the slice elements",
			// value: []any{42, "not a number", 43},
			// targetStruct: &struct {
			// Attr1 []int `some_tag:"attr1"`
			// }{},
			// expectErrToContain: []string{"error decoding field", "Attr1", "int", "string"},
			// },
			{
				desc:  "should report error if tag has no name",
				value: "example-value",
				targetStruct: &struct {
					Attr1 string `valid:"attr1" :"missing_name"`
				}{},
				expectErrToContain: []string{"malformed tag", `valid:"attr1" :"missing_name"`},
			},
			{
				desc:  "should report error if tag has no value",
				value: "example-value",
				targetStruct: &struct {
					Attr1 string `valid:"attr1" missing_value:`
				}{},
				expectErrToContain: []string{"malformed tag", `valid:"attr1" missing_value:`},
			},
			{
				desc:  "should report error if tag has invalid character",
				value: "example-value",
				targetStruct: &struct {
					Attr1 string `line_break
												"attr1"`
				}{},
				// (10 is the ascii number for line breaks)
				expectErrToContain: []string{"malformed tag", "10"},
			},
			{
				desc:  "should report error if tag value is missing quotes",
				value: "example-value",
				targetStruct: &struct {
					Attr1 string `line_break:attr1"`
				}{},
				expectErrToContain: []string{"malformed tag", "missing quotes", `line_break:attr1"`},
			},
			{
				desc:  "should report error if tag value is missing quotes",
				value: "example-value",
				targetStruct: &struct {
					Attr1 string `line_break:"attr1`
				}{},
				expectErrToContain: []string{"malformed tag", "missing end quote", `line_break:"attr1`},
			},
		}
		for _, test := range tests {
			t.Run(test.desc, func(t *testing.T) {
				err := structi.ForEach(test.targetStruct, func(field structi.Field) error {
					return field.Set(test.value)
				})
				tt.AssertErrContains(t, err, test.expectErrToContain...)
			})
		}
	})

	t.Run("wrap errors correctly", func(t *testing.T) {

		t.Run("wrap error from Decoder", func(t *testing.T) {
			// Use a int parse error as the type example
			err := structi.ForEach(
				&struct {
					A int `a:""`
				}{},
				func(field structi.Field) error {
					_, err := strconv.ParseInt("not-an-int", 10, 0)
					return err
				},
			)

			var parseErr *strconv.NumError
			tt.AssertTrue(t, errors.As(err, &parseErr), "error %#v should wrap %T", err, parseErr)
			tt.AssertEqual(t, parseErr.Err, strconv.ErrSyntax)
		})

		t.Run("wrap error from slice conversion", func(t *testing.T) {
			// Use a int parse error as the type example
			err := structi.ForEach(
				&struct {
					A []int
				}{},
				func(field structi.Field) error {
					return field.Set([]string{"not-an-int"})
				},
			)

			// Sanity check: the outer error _does_ contain the string we don't want to see in the wrapped error
			tt.AssertErrContains(t, err, "iteration error", "A", "string", "[]int", "not-an-int")

			// In this case, it should just be a wrapped sting error
			wrapped := errors.Unwrap(err)
			tt.AssertNotEqual(t, wrapped, nil)
		})
	})
}

func TestGetStructInfo(t *testing.T) {
	type MyStruct struct {
		A int
	}

	t.Run("should work for a struct pointer", func(t *testing.T) {
		var s MyStruct
		si, err := structi.GetStructInfo(&s)
		tt.AssertNoErr(t, err)
		tt.AssertEqual(t, len(si.Fields), 1)
	})

	t.Run("should fail for a struct", func(t *testing.T) {
		var s MyStruct
		_, err := structi.GetStructInfo(s)
		tt.AssertErrContains(t, err, "struct pointer", "MyStruct")
	})

	t.Run("should work for reflect.Type", func(t *testing.T) {
		typ := reflect.TypeOf(MyStruct{})
		si, err := structi.GetStructInfo(typ)
		tt.AssertNoErr(t, err)
		tt.AssertEqual(t, len(si.Fields), 1)
	})
	t.Run("should work for reflect.Type of a struct pointer", func(t *testing.T) {
		typ := reflect.TypeOf(&MyStruct{})
		si, err := structi.GetStructInfo(typ)
		tt.AssertNoErr(t, err)
		tt.AssertEqual(t, len(si.Fields), 1)
	})
	t.Run("should fail if input type is not a kind of struct or struct pointer", func(t *testing.T) {
		typ := reflect.TypeOf(1)
		_, err := structi.GetStructInfo(typ)
		tt.AssertErrContains(t, err, "can only get struct info from structs", "int")
	})
}

func intPtr(i int) *int {
	return &i
}
