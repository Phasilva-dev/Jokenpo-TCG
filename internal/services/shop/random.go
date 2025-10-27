//START OF FILE jokenpo/internal/services/shop/random.go
package shop

import (
	"jokenpo/internal/game/card"
	"math/rand/v2"
)

func generateRandomCardTypo(r *rand.Rand) string {
	randomIndexType := r.IntN(len(card.CardTypes))
	randomType := card.CardTypes[randomIndexType]
	return randomType
}

func generateRandomCardColor(r *rand.Rand) string {
	randomIndexColor := r.IntN(len(card.CardColors))
	randomColor := card.CardColors[randomIndexColor]
	return randomColor
}

// WeightedValue (struct inalterada)
type weightedValue struct {
	Value  uint8
	Weight int
}

// valueDistribution agora favorece valores altos.
// A soma dos pesos ainda é 100.
var valueDistribution = []weightedValue{
	{Value: 1, Weight: 14},   
	{Value: 2, Weight: 14},   
	{Value: 3, Weight: 14},   
	{Value: 4, Weight: 14},   
	{Value: 5, Weight: 12},   
	{Value: 6, Weight: 10},  
	{Value: 7, Weight: 7},  
	{Value: 8, Weight: 6},  
	{Value: 9, Weight: 5},  
	{Value: 10, Weight: 4}, 
}

// totalWeight (variável inalterada)
var totalWeight int

// A função init() recalcula o peso total automaticamente.
// NENHUMA MUDANÇA É NECESSÁRIA AQUI.
func init() {
	totalWeight = 0
	for _, wv := range valueDistribution {
		totalWeight += wv.Weight
	}
}

func generateRandomCardValue(r *rand.Rand) uint8 {
	roll := r.IntN(totalWeight)
	for _, wv := range valueDistribution {
		roll -= wv.Weight
		if roll < 0 {
			return wv.Value
		}
	}
	return valueDistribution[len(valueDistribution)-1].Value
}

//END OF FILE jokenpo/internal/services/shop/random.go