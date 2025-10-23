//START OF FILE jokenpo/internal/game/card/card.go
package card

import (
	
)

type Card struct {
	typo string
	value uint8
	color string
}



func (c *Card) Typo() string  { return c.typo }
func (c *Card) Value() uint8  { return c.value }
func (c *Card) Color() string { return c.color }

func (c *Card) Key() string { return CardKey(c.typo,c.value,c.color) }



// ---- Construtor ----

func newCard(typo string, value uint8, color string) (*Card, error) {
	card := &Card{typo: typo, value: value, color: color}

	validators := []cardValidator{
		validateTypo,
		validateValue,
		validateColor,
	}

	for _, v := range validators {
		if err := v(card); err != nil {
			return nil, err
		}
	}

	return card, nil
}

func (c *Card) String() string {
	return CardKey(c.typo,c.value,c.color)
}

//END OF FILE jokenpo/internal/game/card/card.go