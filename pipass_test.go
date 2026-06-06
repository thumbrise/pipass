//go:build test

package pipass_test

import (
	"testing"

	"github.com/thumbrise/pipass/testdata/generated"
)

// MockLedger implements pipass.Ledger to collect transaction traces in memory
type MockLedger struct {
	paths []string
}

func (m *MockLedger) Log(path string, reason string, prev, next any) {
	m.paths = append(m.paths, path)
}

func TestPipassComplexHierarchy(t *testing.T) {
	ledger := &MockLedger{}

	// 1. Initialize the pipeline root (Stage)
	stage := generated.NewStagePipePass("stage", ledger)
	stage.SetTitle("Act I: The Discovery", "initializing stage title")
	stage.SetTitle("Act I: The Discovery", "duplicate setting for idempotency validation")

	if len(ledger.paths) != 1 || ledger.paths[0] != "stage.Title" {
		t.Errorf("expected scalar trace 'stage.Title', got: %v", ledger.paths)
	}

	// 2. Validate multiple parallel rails (Actors and Triggers) on the same level
	actor := generated.NewActorPipePass("", nil)
	actor.SetName("Sherlock", "naming the main character")
	stage.AppendActors(actor, "introducing an actor to the scene")

	trigger := generated.NewTriggerPipePass("", nil)
	trigger.SetEvent("OnEnter", "defining trigger condition")
	stage.AppendTriggers(trigger, "attaching a trigger to the stage")

	if len(ledger.paths) < 3 || ledger.paths[1] != "stage.Actors[0]" || ledger.paths[2] != "stage.Triggers[0]" {
		t.Errorf("parallel rail tracking indices calculated incorrectly: %v", ledger.paths)
	}

	// 3. Validate context path propagation inside independent Map traversals
	err := stage.MapActors(func(a generated.ActorPass) error {
		a.SetRole("Detective", "assigning role via map traversal")
		return nil
	})
	if err != nil {
		t.Fatalf("MapActors failed: %v", err)
	}

	if len(ledger.paths) < 4 || ledger.paths[3] != "stage.Actors[0].Role" {
		t.Errorf("expected contextual path 'stage.Actors[0].Role', got: %v", ledger.paths)
	}

	// 4. Validate deep self-recursion inside Triggers
	err = stage.MapTriggers(func(parent generated.TriggerPass) error {
		childTrigger := generated.NewTriggerPipePass("", nil)
		childTrigger.SetEvent("OnInteract", "nested event definition")

		// Cast to underlying engine structure for internal Append operations
		if p, ok := parent.(*generated.TriggerPipePass); ok {
			p.AppendChildren(childTrigger, "birthing a deep nested sub-trigger")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("MapTriggers failed: %v", err)
	}

	if len(ledger.paths) < 5 || ledger.paths[4] != "stage.Triggers[0].Children[0]" {
		t.Errorf("recursive child tracking index or path calculated incorrectly: %v", ledger.paths)
	}
}
