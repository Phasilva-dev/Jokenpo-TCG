package deck

import (
	"jokenpo/internal/game/card"
	"math/rand/v2"

	"fmt"
)

type pileOfCards []*card.Card

func (p *pileOfCards) Shuffle(r *rand.Rand) {
	n := len(*p)
	for i := n - 1; i > 0; i-- {
		j := r.IntN(i + 1)
		(*p)[i], (*p)[j] = (*p)[j], (*p)[i]
	}
}

func (p *pileOfCards) GetCard(index int) (*card.Card, error) {
	n := len(*p)
	if index < 0 || index >= n {
		return nil, fmt.Errorf("index %d out of range", index)
	}
	return (*p)[index], nil
}

func (p *pileOfCards) DrawCard(r *rand.Rand) (*card.Card, error) {
	n := len(*p)
	if n == 0 {
		return nil, fmt.Errorf("pile is empty")
	}
	index := r.IntN(n)
	card := (*p)[index]

	// remove a carta do slice
	*p = append((*p)[:index], (*p)[index+1:]...)
	return card, nil
}

func (p *pileOfCards) DrawTop() (*card.Card, error) {
	n := len(*p)
	if n == 0 {
		return nil, fmt.Errorf("pile is empty")
	}

	top := (*p)[0]        // carta do topo
	*p = (*p)[1:]          // remove do slice
	return top, nil
}

func (p *pileOfCards) AddCard(c *card.Card) {
	*p = append(*p, c)
}

func (p *pileOfCards) RemoveCard(c *card.Card) error {
	for i, card := range *p {
		if card == c { // compara ponteiros, já que as cartas globais são imutáveis
			*p = append((*p)[:i], (*p)[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("card not found in pile")
}

