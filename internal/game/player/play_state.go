//START OF FILE jokenpo/internal/game/player/play_state.go
package player

import (
	"fmt"
	"jokenpo/internal/game/card"
	"jokenpo/internal/game/deck"
	"math/rand/v2"
)



func (p *Player) EndPlay() (error) {
	if p.state != PLAY {
		return fmt.Errorf("error: Player must be in PLAY state to end a play")
	}
	p.inventory.GameDeck().ResetToDeck()
	p.state = MENU
	return nil

}

func (p *Player) DrawToHand() (*card.Card, error) {
	if p.state != PLAY {
		return nil, fmt.Errorf("error: Player must be in PLAY state to draw a card to hand")
	}
	card, err := p.Inventory().GameDeck().DrawToHand()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func (p *Player) PlayCardFromHand(index int) (*card.Card, error) {
	if p.state != PLAY {
		return nil, fmt.Errorf("error: Player must be in PLAY state to play a card from hand")
	}
	card, err := p.Inventory().GameDeck().PlayCardFromHand(index)
	if err != nil {
		return nil, err
	}
	return card, nil
}

func (p *Player) PlayRandomCardFromHand(r *rand.Rand) (*card.Card, error) {
	if p.state != PLAY {
		return nil, fmt.Errorf("error: Player must be in PLAY state to play a random card to hand")
	}
	card, err := p.Inventory().GameDeck().PlayRandomCardFromHand(r)
	if err != nil {
		return nil, err
	}
	return card, nil
}

func (p *Player) ResolvePlay(won bool) (string, error) {
	if p.state != PLAY {
		return "", fmt.Errorf("error: Player must be in PLAY state to resolve play")
	}
	return p.inventory.GameDeck().ResolvePlay(won)

}

func (p *Player) SeeHand() (string, error) {
	if p.state != PLAY {
		return "", fmt.Errorf("error: Player must be in PLAY state to see hand")
	}
	return p.inventory.GameDeck().ZoneString(deck.HAND)
}

func (p *Player) WinCondition() (bool, error) {
	if p.state != PLAY {
		return false, fmt.Errorf("error: Player must be in PLAY state to check win condition")
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
//END OF FILE jokenpo/internal/game/player/play_state.go