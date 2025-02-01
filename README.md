[![CI](https://github.com/VinGarcia/structscanner/actions/workflows/ci.yml/badge.svg)](https://github.com/VinGarcia/structscanner/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/VinGarcia/structscanner/branch/master/graph/badge.svg?token=5CNJ867C66)](https://codecov.io/gh/VinGarcia/structscanner)
[![Go Reference](https://pkg.go.dev/badge/github.com/vingarcia/structi.svg)](https://pkg.go.dev/github.com/vingarcia/structi)
![Go Report Card](https://goreportcard.com/badge/github.com/vingarcia/structi)

# Welcome to the Go Struct Iterator

This project was created to make it safe, easy and efficient
to use reflection to read and write data to and from structs.

Note that it is always faster if you don't use reflection,
but when you need to use it the most efficient way of doing it
is by caching the info you get from the types which is
something this library do.

So to make it clear, this library is not something like:

- https://github.com/mitchellh/mapstructure

Nor something like:

https://github.com/spf13/viper

This is a library for allowing you to write your own Viper
or Mapstructure libraries with ease and in a few lines of code,
so that you get exactly what you need and in the way you need it.

So the examples below are examples of things you can get by using
this library. Both examples are also public so you can use them
directly if you want.

But the interesting part is that both were written
in very few lines of code.

## Usage Examples:

### Loading data from `os.Getenv()`:

The code below will fill the struct with data from env variables.

It will use the `env` tags to map which env var should be used
as source for each of the attributes of the struct.

```golang
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/vingarcia/structi"
)

func main() {
	var config struct {
		GoPath     string `env:"GOPATH"`
		Home       string `env:"HOME"`
		CurrentDir string `env:"PWD"`
		Shell      string `env:"SHELL"`
	}

	err := structi.ForEach(&config, func(field structi.Field) error {
		envTag := field.Tags["env"]
		if envTag != "" {
			return field.Set(os.Getenv(envTag))
		}
		return nil
	})
	if err != nil {
		log.Fatalf("error loading env vars: %v", err)
	}

	b, _ := json.MarshalIndent(config, "", "  ")
	fmt.Println("loaded config:", string(b))
}
```

### Loading data from a map:

This second example will fill a struct with the values of an input map, and it will
also handle nested substructs using recursion:

```golang
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"

	"github.com/vingarcia/structi"
)

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

// This main func just illustrates the usage of the LoadFromMap function above
func main() {
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
```

## What info can I get from each attribute of the struct?

> Note that the actual struct is slightly different, it is shown like this for simplicity

```golang
type Field struct {
	Tags map[string]string
	Name string
	Kind reflect.Kind
	Type reflect.Type

	IsEmbeded bool

	Set   func(value any) error
	Value any
}
```

## GetStructInfo function

If you wish to use the Field info (names, tags, type etc) elsewhere you can use the `GetStructInfo()` function.

```golang
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
```

It is possible to pass a `reflection.Type` object to `GetStructInfo`, which is particularly useful for nested structs:

```golang
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
```


## License

This project was put into public domain, which means you can copy, use and modify
any part of it without mentioning its original source so feel free to do that
if it would be more convenient that way.

Enjoy.
