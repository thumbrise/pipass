package userland

import (
	"log/slog"
	"time"
)

type Person struct {
	logger    *slog.Logger
	email     string
	birthdate time.Time
	friends   []Person
	meta      map[string]any
	mad       bool
}
