package player

import (
	"fmt"
	"jokenpo/internal/game/player/inventory"
	"math/rand/v2"
)

const PLAY = "playing"
const MENU = "menu"

var possibleStates = []string{PLAY, MENU}

type Player struct {
	inventory *inventory.Inventory
	state string
	rng *rand.Rand
}

func NewPlayer(seed uint64) *Player {
	// PCG é um gerador rápido e de boa qualidade para jogos.
	rngSource := rand.NewPCG(seed, 1) 

	return &Player{
		inventory: inventory.NewInventory(),
		state:     "menu",
		rng:       rand.New(rngSource), // Cria a instância do Rand a partir da fonte
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



