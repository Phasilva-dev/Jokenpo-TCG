package player

import (
	"fmt"
	"jokenpo/internal/game/deck"
)

func (p *Player) DrawToHand() (string, error) {
	if p.state != PLAY {
		return "", fmt.Errorf("")
	}
	card, err := p.Inventory().GameDeck().DrawToHand()
	if err != nil {
		return "", err
	}
	//String personalizada, algo como "Você comprou a seguinte carta card.string()"
	return card.String(), nil
}

func (p *Player) PlayCardFromHand(index int) (string, error) {
	if p.state != PLAY {
		return "", fmt.Errorf("")
	}
	card, err := p.Inventory().GameDeck().PlayCardFromHand(index)
	if err != nil {
		return "", err
	}
	//String personalizada, algo como "Você comprou a seguinte carta card.string()"
	return card.String(), nil
}

func (p *Player) ResolvePlay(won bool) (string, error) {
	if p.state != PLAY {
		return "", fmt.Errorf("")
	}
	return p.inventory.GameDeck().ResolvePlay(won)

}

func (p *Player) SeeHand() (string, error) {
	if p.state != PLAY {
		return "", fmt.Errorf("")
	}
	return p.inventory.GameDeck().ZoneString(deck.HAND)
}



