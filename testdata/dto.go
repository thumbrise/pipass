package testdata

import "time"

// Config represents Just Data (Black Box / Scalar).
// It is NOT registered in the pipeline node registry.
type Config struct {
	Scope     string
	Variables map[string]interface{}
}

// Stage represents Pipeline Node Level 1 (The Pipeline Root).
type Stage struct {
	ID        string
	Title     string
	Actors    []Actor   // Rail #1: Collection of Actor nodes (Level 2)
	Triggers  []Trigger // Rail #2: Collection of Trigger nodes (Level 2)
	Settings  Config    // Just Data (Scalar struct payload)
	IsActive  bool
	CreatedAt time.Time // Just Data (Scalar)
}

// Actor represents Pipeline Node Level 2 (Nested within Stage).
type Actor struct {
	Name      string
	Role      string
	Inventory map[string]interface{} // Just Data (Map inside a node)
}

// Trigger represents Pipeline Node Level 2 (Nested within Stage).
// Demonstrates deep self-recursion and slice-of-primitives coexistence.
type Trigger struct {
	Event    string
	Actions  []string  // Just Data (Slice of primitive strings)
	Children []Trigger // Deep Recursion: Trigger births nested Triggers
}
