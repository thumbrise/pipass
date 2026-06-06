package main

import (
	"fmt"

	"github.com/thumbrise/pipass"
	"github.com/thumbrise/pipass/examples/game/generated"
)

func main() {
	fmt.Println("=== Game Session Orchestrator ===")

	ledger := &pipass.PrintLedger{}
	session := generated.NewSessionPipePass("session", ledger)

	// 1. Singular node: Config with type-preserved fields
	cfg := generated.NewGameConfigPipePass("", nil)
	cfg.SetMode("adventure", "set game mode")
	cfg.SetTickRate(20.0, "configure tick rate")
	cfg.SetCheats(new(true), "enable debug mode")
	session.SetConfig(cfg, "attach config to session")

	// 2. Pointer to primitive
	session.SetActive(new(true), "session started")
	session.SetMaxPlayers(new(10), "set max players")

	// 3. Slice node with nested singular node
	stats := generated.NewPlayerStatsPipePass("", nil)
	stats.SetHP(100, "init HP")
	stats.SetMP(50, "init MP")
	stats.SetBuffs([]string{"swiftness", "strength"}, "starting buffs")

	alice := generated.NewPlayerPipePass("", nil)
	alice.SetName("Alice", "player joined")
	alice.SetClass(new("warrior"), "choose class")
	alice.SetScore(1500, "starting score")
	alice.SetStats(stats, "attach stats")
	session.AppendPlayers(alice, "add first player")

	// 4. Map with concrete value type
	session.SetMetadataKey("region", "eu-west", "set region")
	session.SetMetadataKey("version", "1.0", "set version")

	// 5. Deep traversal via MapPlayers
	_ = session.MapPlayers(func(p generated.PlayerPass) error {
		p.SetScore(p.Score()+500, "level bonus")
		return nil
	})

	fmt.Println("=== Session created successfully ===")
}
