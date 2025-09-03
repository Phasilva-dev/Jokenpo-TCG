package player

import (
	"fmt"
	"jokenpo/internal/game/card"
	"jokenpo/internal/game/deck"
)



func (p *Player) EndPlay() (error) {
	if p.state != MENU {
		return fmt.Errorf("")
	}
	p.inventory.GameDeck().ResetToDeck()
	p.state = MENU
	return nil

}

func (p *Player) DrawToHand() (*card.Card, error) {
	if p.state != PLAY {
		return nil, fmt.Errorf("")
	}
	card, err := p.Inventory().GameDeck().DrawToHand()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func (p *Player) PlayCardFromHand(index int) (*card.Card, error) {
	if p.state != PLAY {
		return nil, fmt.Errorf("")
	}
	card, err := p.Inventory().GameDeck().PlayCardFromHand(index)
	if err != nil {
		return nil, err
	}
	return card, nil
}

func (p *Player) PlayRandomCardFromHand() (*card.Card, error) {
	if p.state != PLAY {
		return nil, fmt.Errorf("")
	}
	card, err := p.Inventory().GameDeck().PlayRandomCardFromHand(p.rng)
	if err != nil {
		return nil, err
	}
	return card, nil
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

func (p *Player) WinCondition() (bool, error) {
	if p.state != PLAY {
		return false, fmt.Errorf("")
	}
	return p.inventory.GameDeck().WinCondition(), nil
}

// HasNoMoreMoves verifica se o jogador não tem mais cartas no deck nem na mão.
func (p *Player) HasNoMoreMoves() (bool, error) {
	deck, err := p.Inventory().GameDeck().GetCardsInZone("deck")
	if err != nil {
		return false, err // Erro interno se a zona não existir
	}
	hand, err := p.Inventory().GameDeck().GetCardsInZone("hand")
	if err != nil {
		return false, err // Erro interno se a zona não existir
	}
	
	return len(deck) == 0 && len(hand) == 0, nil
}

// CardsInWinPile retorna a contagem de cartas na pilha de vitória.
func (p *Player) CardsInWinPile() (int, error) {
    winPile, err := p.Inventory().GameDeck().GetCardsInZone("win")
    if err != nil {
        return 0, err
    }
    return len(winPile), nil
}




