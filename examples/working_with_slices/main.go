package main

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"

	"github.com/vingarcia/structi"
	"github.com/vingarcia/structi/slicei"
)

func main() {
	var output struct {
		NotASlice       int
		EmptySlice      []uint   `tag:"s1"`
		SliceWithValues []string `tag:"s2"`
	}
	output.SliceWithValues = []string{"foo", "bar"}

	err := structi.ForEach(&output, func(field structi.Field) error {
		if field.Kind != reflect.Slice {
			// Let's ignore non slices for this example
			return nil
		}

		if field.Tags["tag"] == "s1" {
			// 42 is an int and will be converted to uint automatically:
			return slicei.Append(field.Value, int(42))
		}

		return slicei.ForEach(field.Value, func(field slicei.Field) error {
			// Add the index number at the end of each slice value:
			return field.Set(fmt.Sprint(reflect.ValueOf(field.Value).Elem().Interface(), field.Index))
		})
	})
	if err != nil {
		log.Fatalf("error modifying struct: %v", err)
	}

	b, _ := json.MarshalIndent(output, "", "  ")
	fmt.Println("modified struct with slices:", string(b))
}
