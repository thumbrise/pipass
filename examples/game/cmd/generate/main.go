package main

import (
	"log"
	"os"

	"github.com/thumbrise/pipass"
	"github.com/thumbrise/pipass/examples/game/userland"
)

func main() {
	out, err := pipass.Compile(
		"generated",
		userland.Session{},
		userland.GameConfig{},
		userland.Player{},
		userland.PlayerStats{},
	)
	if err != nil {
		log.Fatalf("Compile: %v", err)
	}

	err = os.MkdirAll("generated", 0755)
	if err != nil {
		log.Fatalf("MkdirAll: %v", err)
	}

	err = os.WriteFile("generated/generated.go", out, 0644)
	if err != nil {
		log.Fatalf("WriteFile: %v", err)
	}
}
