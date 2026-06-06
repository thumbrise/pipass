package main

import (
	"fmt"
	"log"

	"github.com/thumbrise/pipass"
	"github.com/thumbrise/pipass/examples/poc/generated"
)

func main() {
	fmt.Println("--- Pipass Pipeline Start ---")

	// 1. Initialize ledger
	ledger := &pipass.PrintLedger{}

	// 2. Create root node of the shadow model
	root := generated.NewPersonPipePass("root", ledger)

	// 3. Plugin #1: Populate root base data
	root.SetEmail("holmes@detective.co", "initialize profile")
	root.SetMad(true, "ontology realization")

	// 4. Plugin #2: Birth of child nodes (graph branching)
	friend := generated.NewPersonPipePass("", nil) // empty for now, Append will compute coordinates
	friend.SetEmail("watson@doctor.co", "basic friend email")

	root.AppendFriends(friend, "add loyal companion")

	// 5. Plugin #3: Graph traversal and pet creation
	if err := root.MapFriends(func(f generated.PersonPass) error {
		pet := generated.NewPetPipePass("", nil)
		pet.SetName("Toby", "hound nickname")

		f.AppendPets(pet, "buy a dog for the doctor")
		return nil
	}); err != nil {
		log.Fatalf("MapFriends traversal failed: %v", err)
	}

	// 6. Plugin #4: Deep mutation via computed path
	if err := root.MapFriends(func(f generated.PersonPass) error {
		return f.MapPets(func(p generated.PetPass) error {
			p.SetName("Sherlock-Junior", "dog rebranding")
			return nil
		})
	}); err != nil {
		log.Fatalf("MapFriends deep mutation failed: %v", err)
	}

	fmt.Println("--- Pipeline completed successfully ---")
}
