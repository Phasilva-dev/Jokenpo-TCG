package card

import (
	"math/rand/v2"
	"strings"
	"fmt"
)

type Pile []*Card

// Size retorna o número de cartas na pilha.
func (p *Pile) Size() int {
	// Adicionamos uma verificação para o caso de o ponteiro *Pile ser nil.
	// Isso torna o método mais seguro de usar.
	if p == nil {
		return 0
	}
	return len(*p)
}

func (p *Pile) Shuffle(r *rand.Rand) {
	n := p.Size()
	if n > 1 {
		for i := n - 1; i > 0; i-- {
			j := r.IntN(i + 1)
			(*p)[i], (*p)[j] = (*p)[j], (*p)[i]
		}
	}
}

func (p *Pile) GetCard(index int) (*Card, error) {
	if index < 0 || index >= p.Size() {
		return nil, fmt.Errorf("index %d out of range", index)
	}
	return (*p)[index], nil
}

func (p *Pile) DrawCard(r *rand.Rand) (*Card, error) {
	n := p.Size()
	if n == 0 {
		return nil, fmt.Errorf("pile is empty")
	}
	index := r.IntN(n)
	card := (*p)[index]

	// remove a carta do slice
	*p = append((*p)[:index], (*p)[index+1:]...)
	return card, nil
}

func (p *Pile) DrawTop() (*Card, error) {
	if p.Size() == 0 {
		return nil, fmt.Errorf("pile is empty")
	}

	top := (*p)[0]  // carta do topo
	*p = (*p)[1:]    // remove do slice
	return top, nil
}

func (p *Pile) AddCard(c *Card) {
	*p = append(*p, c)
}

func (p *Pile) RemoveCard(c *Card) error {
	for i, card := range *p {
		if card == c { // compara ponteiros, já que as cartas globais são imutáveis
			*p = append((*p)[:i], (*p)[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("card not found in pile")
}

// RemoveCardByIndex removes a card from the pile at a specific index.
// It modifies the original slice.
func (p *Pile) RemoveCardByIndex(index int) (*Card, error) {
	if index < 0 || index >= p.Size() {
		return nil, fmt.Errorf("index %d is out of bounds for a pile of size %d", index, p.Size())
	}
	
	card := (*p)[index] // Pega a carta antes de remover

	*p = append((*p)[:index], (*p)[index+1:]...)

	return card, nil
}



func (p *Pile) String() string {
	if p == nil || p.Size() == 0 {
		return "(Empty)"
	}

	var sb strings.Builder
	sb.WriteString("--------------------\n")

	for i, c := range *p {
		if c == nil {
			sb.WriteString(fmt.Sprintf("[%d]: <nil card>\n", i))
		} else {
			// índice do slice + representação da carta
			sb.WriteString(fmt.Sprintf("[%d]: %s\n", i, c))
		}
	}

	sb.WriteString("--------------------")
	return sb.String()
}

