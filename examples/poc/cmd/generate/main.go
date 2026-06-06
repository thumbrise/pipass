package main

import (
	"log"
	"os"

	"github.com/thumbrise/pipass/examples/poc/userland"
)

import "github.com/thumbrise/pipass"

func main() {
	out, err := pipass.Compile("generated", userland.Person{}, userland.Pet{})
	if err != nil {
		log.Fatalf("Compile: %v", err)
	}

	err = os.MkdirAll("generated", 0755)
	if err != nil {
		log.Fatalf("MkdirAll: %v", err)
	}

	err = os.WriteFile("generated/generated.go", out, 0600)
	if err != nil {
		log.Fatalf("WriteFile: %v", err)
	}
}
