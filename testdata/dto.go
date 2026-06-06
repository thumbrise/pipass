package testdata

import "time"

// Custom named types for type-preservation verification.
// These exercise the generator's ability to emit the original type
// instead of falling back to interface{}.
type Score float64

type Template struct {
	Name    string
	Version int
}

// Config represents Just Data (Black Box / Scalar).
// It is NOT registered in the pipeline node registry.
type Config struct {
	Scope     string
	Variables map[string]interface{}
}

// Stage represents Pipeline Node Level 1 (The Pipeline Root).
type Stage struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Actors    []Actor   // Rail #1: Collection of Actor nodes (Level 2)
	Triggers  []Trigger // Rail #2: Collection of Trigger nodes (Level 2)
	Settings  Config    // Just Data (Scalar struct payload)
	IsActive  bool
	CreatedAt time.Time // Just Data (Scalar)

	// Type-preservation test cases
	Score    Score      // named type based on float64
	Ratio    *float64   // pointer to primitive
	Template *Template  // pointer to named struct
}

// Actor represents Pipeline Node Level 2 (Nested within Stage).
type Actor struct {
	Name      string
	Role      string
	Inventory map[string]interface{} // Just Data (Map inside a node)

	// Type-preservation test cases
	Nickname *string // pointer to string
	IsHero   *bool   // pointer to bool
}

// Trigger represents Pipeline Node Level 2 (Nested within Stage).
// Demonstrates deep self-recursion and slice-of-primitives coexistence.
type Trigger struct {
	Event    string
	Actions  []string  // Just Data (Slice of primitive strings)
	Children []Trigger // Deep Recursion: Trigger births nested Triggers

	// Type-preservation test cases
	Metadata map[string]string // Map with concrete value type
	Priority *int              // pointer to int
	Index    map[int]string    // Map with non-string key type
}
