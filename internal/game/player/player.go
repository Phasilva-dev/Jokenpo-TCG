package player

import (
	"jokenpo/internal/game/card"
	"jokenpo/internal/game/deck"
	//"jokenpo/internal/network"
)

type Player struct {
	collection *card.PlayerCollection
	deck *deck.Deck
	playing bool
	//client *network.Client jogar isso pro NetWorkPlayer
}

func NewPlayer() *Player {
	return &Player{
		collection: card.NewPlayerCollection(),
		deck: deck.NewDeck(),
		playing: false,
	}
}

func (p *Player) SeeDeck() {
	p.deck.PrintZone("deck")
}
