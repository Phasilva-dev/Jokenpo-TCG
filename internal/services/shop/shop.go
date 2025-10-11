package shop

import (
	"fmt"
	"jokenpo/internal/game/card"
	"math"
	"math/rand/v2"
	"time"
)

type Shop struct {
	packageCount uint64
	rng          *rand.Rand
}

func NewShop() *Shop {
	seed := uint64(time.Now().UnixNano())
	return &Shop{
		packageCount: 0,
		rng:          rand.New(rand.NewPCG(seed, 0)),
	}
}

const maxPurchases = math.MaxUint64
const packageSize = 3

func (s *Shop) purchasePackage(quantity uint64) ([]*card.Card, error) {

	if quantity == 0 {
		return nil, fmt.Errorf("invalid quantity: must be greater than zero")
	}

	if s.packageCount + quantity >= maxPurchases {
		return nil, fmt.Errorf("cannot process purchase: maximum purchase limit reached")
	}

	totalCards := int(quantity) * packageSize
	allCards := make([]*card.Card, 0, totalCards)

	for i := uint64(0); i < quantity; i++ {
		for j := 0; j < packageSize; i++ {
			typo := generateRandomCardTypo(s.rng)
			value := generateRandomCardValue(s.rng)
			color := generateRandomCardColor(s.rng)
			key := card.CardKey(typo, value, color)

			c, err := card.GetCard(key)
			if err != nil {
				return nil, fmt.Errorf("failed to generate valid card for package %d, card %d: %w", i+1, j+1, err)
			}

			allCards = append(allCards, c)
		}
	}

	s.packageCount += quantity
	return allCards, nil

}