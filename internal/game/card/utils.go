//START OF FILE jokenpo/internal/game/card/utils.go
package card

import (
	"fmt"
	"strings"
)

func CardKey (typo string, value uint8, color string) string{
	return fmt.Sprintf("%s:%d:%s", typo, value, color)
}

// Tipo para funções de validação
type cardValidator func(*Card) error

// Lista de tipos permitidos
var allowedTypes = map[string]struct{}{
	"rock":    {},
	"paper":   {},
	"scissor": {},
}

// Lista de cores permitidas
var allowedColors = map[string]struct{}{
	"red":   {},
	"blue":  {},
	"green": {},
}

var (
	CardTypes  = []string{"rock", "paper", "scissor"}
	CardColors = []string{"red", "green", "blue"}
)

// ---- Funções de validação ----

func validateTypo(c *Card) error {
	if _, ok := allowedTypes[c.typo]; !ok {
		return fmt.Errorf("invalid card type: %s", c.typo)
	}
	return nil
}

func validateValue(c *Card) error {
	if c.value == 0 || c.value > 10 {
		return fmt.Errorf("invalid card value: %d (must be 1–10)", c.value)
	}
	return nil
}

func validateColor(c *Card) error {
	if _, ok := allowedColors[c.color]; !ok {
		return fmt.Errorf("invalid card color: %s", c.color)
	}
	return nil
}

func SliceOfCardsToString(cards []*Card) string {

	if cards == nil || len(cards) == 0 {
		return "(No cards to display)"
	}

	var sb strings.Builder

	for i, c := range cards {
		line := fmt.Sprintf("[%d]: %s\n", i, c)

		sb.WriteString(line)
	}
	return strings.TrimSuffix(sb.String(), "\n")
}

//END OF FILE jokenpo/internal/game/card/utils.go