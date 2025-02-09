package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"

	"github.com/vingarcia/structi"
)

func main() {
	var output struct {
		Attr1       int `env:"attr1"`
		OtherStruct *struct {
			Attr2 int `env:"attr2"`
		}
	}
	err := structi.ForEach(&output, func(field structi.Field) error {
		if field.Kind == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct {
			subStruct := reflect.New(field.Type.Elem())

			return errors.Join(
				structi.ForEach(subStruct, func(field structi.Field) error {
					return field.Set(42)
				}),
				field.Set(subStruct),
			)
		}

		return field.Set(64)
	})
	if err != nil {
		log.Fatalf("error modifying struct: %v", err)
	}

	b, _ := json.MarshalIndent(output, "", "  ")
	fmt.Println("modified struct:", string(b))
}
