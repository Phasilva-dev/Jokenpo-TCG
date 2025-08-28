package player

import (
	"jokenpo/internal/game/player/inventory"
)

const play = "playing"
const MENU = "menu"

type Player struct {
	inventory *inventory.Inventory
	state string
}

func NewPlayer() *Player {
	return &Player{
		inventory: inventory.NewInventory(),
		state: "in-shop",
	}
}

func (p *Player) Inventory() *inventory.Inventory {
	return p.inventory
}



