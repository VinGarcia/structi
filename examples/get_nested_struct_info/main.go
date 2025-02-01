package main

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/vingarcia/structi"
)

func main() {
	type Address struct {
		Street string `map:"street"`
		City   string `map:"city"`
	}
	type User struct {
		Name    string  `map:"name"`
		HomeDir Address `map:"home"`
	}

	info, err := structi.GetStructInfo(&User{})
	if err != nil {
		panic(err)
	}

	for _, field := range info.Fields {
		b, _ := json.Marshal(field.Tags)
		fmt.Printf("Field %q has tags %v\n", field.Name, string(b))

		if field.Kind == reflect.Struct {
			nestedInfo, err := structi.GetStructInfo(field.Type)
			if err != nil {
				panic(err)
			}

			fmt.Printf("Nested Field %q has %d fields\n", field.Name, len(nestedInfo.Fields))
		}
	}
}
