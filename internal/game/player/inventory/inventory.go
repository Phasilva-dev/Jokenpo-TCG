package inventory

import (
	"jokenpo/internal/game/card"
	"jokenpo/internal/game/deck"
)

type Inventory struct {
	collection *card.PlayerCollection
	gameDeck *deck.Deck
}

func NewInventory() *Inventory {
	return &Inventory{
		collection: card.NewPlayerCollection(),
		gameDeck: deck.NewDeck(),
	}
}

func (i *Inventory) Collection() *card.PlayerCollection { return i.collection }
func (i *Inventory) GameDeck() *deck.Deck { return i.gameDeck }