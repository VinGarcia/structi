package main

import (
	"encoding/json"
	"fmt"

	"github.com/vingarcia/structi"
)

func main() {
	type User struct {
		Name    string `map:"name"`
		HomeDir string `map:"home"`
	}

	info, err := structi.GetStructInfo(&User{})
	if err != nil {
		panic(err)
	}

	for _, field := range info.Fields {
		b, _ := json.Marshal(field.Tags)
		fmt.Printf("Field %q has tags %v\n", field.Name, string(b))
	}
}
