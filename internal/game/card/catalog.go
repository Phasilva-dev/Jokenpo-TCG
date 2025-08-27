package card

import (
	"fmt"
)

var allCards map[string]*Card

func InitGlobalCatalog() error {
	allCards = make(map[string]*Card)

	for t := range allowedTypes {
		for c := range allowedColors {
			for v := 1; v <= 10; v++ {
				card, err := newCard(t, uint8(v), c)
				if err != nil {
					return err
				}
				allCards[CardKey(t, uint8(v), c)] = card
			}
		}
	}
	return nil
}

// acesso público ao catálogo
func GetCard(key string) (*Card, error) {
	//key := CardKey(typo, value, color) typo string, value uint8, color string
	if card, ok := allCards[key]; ok {
		return card, nil
	}
	return nil, fmt.Errorf("card not found: %s", key)
}