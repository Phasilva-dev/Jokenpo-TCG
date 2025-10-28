//START OF FILE jokenpo/internal/game/player/inventory/inventory.go
package inventory

import (
	"jokenpo/internal/game/card"
	"jokenpo/internal/game/deck"
	"strings"
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

func (i *Inventory) String() string {
	if i == nil {
		return "(Empty Inventory)"
	}

	var sb strings.Builder
	sb.WriteString("\n========== PLAYER INVENTORY ==========\n")

	// Coleção do jogador
	if i.collection != nil {
		//sb.WriteString("Player Collection:\n")
		sb.WriteString(i.collection.String() + "\n")
	} else {
		sb.WriteString("Player Collection: (empty)\n")
	}

	// Deck do jogo
	if i.gameDeck != nil {
		//sb.WriteString("Game Deck:\n")
		sb.WriteString(i.gameDeck.String() + "\n")
	} else {
		sb.WriteString("Game Deck: (empty)\n")
	}

	sb.WriteString("=====================================\n")
	return sb.String()
}
//END OF FILE jokenpo/internal/game/player/inventory/inventory.go