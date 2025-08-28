package player

import (
	"jokenpo/internal/game/player/inventory"
	//"jokenpo/internal/network"
)

type Player struct {
	inventory *inventory.Inventory
	playing bool
	//client *network.Client jogar isso pro NetWorkPlayer
}

func NewPlayer() *Player {
	return &Player{
		inventory: inventory.NewInventory(),
		playing: false,
	}
}

func (p *Player) Inventory() *inventory.Inventory {
	return p.inventory
}

func (p *Player) IsPlaying() bool {
	return p.playing
}

func (p *Player) SeeDeck() {
	p.inventory.GameDeck().PrintZone("deck")
}
