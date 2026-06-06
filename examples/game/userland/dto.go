package userland

import "time"

type Score int64

type GameConfig struct {
	Mode     string
	TickRate float64
	Cheats   *bool
}

type PlayerStats struct {
	HP    int
	MP    int
	Buffs []string
}

type Player struct {
	Name  string
	Score Score
	Class *string
	Stats PlayerStats
}

type Session struct {
	ID         string
	Active     *bool
	MaxPlayers *int
	CreatedAt  time.Time
	Config     GameConfig
	Players    []Player
	Metadata   map[string]string
}
