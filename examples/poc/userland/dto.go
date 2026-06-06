package userland

import (
	"time"
)

type Person struct {
	Email     string
	Birthdate time.Time
	Friends   []Person
	Meta      map[string]any
	Pets      []Pet
	Mad       bool
}

type Pet struct {
	Name       string
	Friends    []Pet
	Owner      Person
	PrevOwners []Person
}
