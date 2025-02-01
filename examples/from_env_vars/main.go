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
		if field.Tags["env"] != "" {
			return field.Set(os.Getenv(field.Tags["env"]))
		}
		return nil
	})
	if err != nil {
		log.Fatalf("error loading env vars: %v", err)
	}

	b, _ := json.MarshalIndent(config, "", "  ")
	fmt.Println("loaded config:", string(b))
}
