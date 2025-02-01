package main

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"

	"github.com/vingarcia/structi"
)

func main() {
	// This one has state and maps a single map to a struct,
	// so you might need to instantiate a new decoder for each input map:
	var user struct {
		ID       int    `map:"id"`
		Username string `map:"username"`
		Address  struct {
			Street  string `map:"street"`
			City    string `map:"city"`
			Country string `map:"country"`
		} `map:"address"`
		SomeSlice []int `map:"some_slice"`
	}

	err := LoadFromMap(&user, map[string]any{
		"id":       42,
		"username": "fakeUsername",
		"address": map[string]interface{}{
			"street":  "fakeStreet",
			"city":    "fakeCity",
			"country": "fakeCountry",
		},
		// Note that even though the type of the slice below
		// differs from the struct slice it will convert all
		// values correctly:
		"some_slice": []float64{1.0, 2.0, 3.0},
	})
	if err != nil {
		log.Fatalf("error loading data from map: %v", err)
	}

	b, _ := json.MarshalIndent(user, "", "  ")
	fmt.Println("loaded user:", string(b))
}

func LoadFromMap(structPtr any, inputMap map[string]any) error {
	return structi.ForEach(structPtr, func(field structi.Field) error {
		tagValue := field.Tags["map"]
		if tagValue == "" {
			return nil
		}

		if field.Kind == reflect.Struct {
			subMap, _ := inputMap[tagValue].(map[string]any)
			if subMap != nil {
				return LoadFromMap(field.Value, subMap)
			}
		}

		return field.Set(inputMap[tagValue])
	})
}
