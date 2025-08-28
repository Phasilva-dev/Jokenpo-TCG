package player

import (
	"jokenpo/internal/game/player/inventory"
	"fmt"
)

const PLAY = "playing"
const MENU = "menu"

var possibleStates = []string{PLAY, MENU}

type Player struct {
	inventory *inventory.Inventory
	state string
}

func NewPlayer() *Player {
	return &Player{
		inventory: inventory.NewInventory(),
		state: "menu",
	}
}

func (p *Player) Inventory() *inventory.Inventory {
	return p.inventory
}

func (p *Player) ChangeState(newState string) error {
	for _, s := range possibleStates {
		if s == newState {
			p.state = newState
			return nil
		}
	}
	return fmt.Errorf("invalid state: %s", newState)
}



