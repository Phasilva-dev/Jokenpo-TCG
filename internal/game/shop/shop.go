package shop

import (
	"fmt"
	"jokenpo/internal/game/card"
	"math"
	"math/rand/v2"
	//"strings"
)

type Shop struct {
	packageCount uint64
}

func NewShop() *Shop {
	return &Shop{
		packageCount: 0,
	}
}

const maxPurchases = math.MaxUint64

func (s *Shop) PurchasePackage(r *rand.Rand, collection *card.PlayerCollection) ([]*card.Card, error) {

	if s.packageCount >= maxPurchases {
		return nil, fmt.Errorf("cannot process purchase: maximum purchase limit reached")
	}

	const packageSize = 3
	keys := make([]string, packageSize)
	cardsToAdd := make([]*card.Card, packageSize)

	for i := 0; i < 3; i++ {
		typo := generateRandomCardTypo(r)
		value := generateRandomCardValue(r)
		color := generateRandomCardColor(r)
		key := card.CardKey(typo,value,color)
		keys[i] = key

		// Validate the generated card key by trying to fetch it from the global catalog.
		// This is the crucial validation step before any modification.
		card, err := card.GetCard(key)
		if err != nil {
			// If even one card is invalid, the entire package purchase fails.
			// This prevents an inconsistent state.
			return nil, fmt.Errorf("failed to generate a valid card for the package: %w", err)
		}
		
		keys[i] = key
		cardsToAdd[i] = card
	}
	/*for i := 0; i < 3; i++ {
		err := collection.AddCard(keys[i],1)
		if err != nil {
			return nil, err
		}
	}*/


	
	

	// --- STAGE 2: EXECUTION ---
	// All cards have been generated and validated. Now, we can safely add them to the collection.
	// This loop is much safer because we know every AddCard call will succeed
	// in terms of card validity.
	for _, key := range keys {
		err := collection.AddCard(key,1)
		if err != nil {
			return nil, err
		}
	}

	// 
	/*var sb strings.Builder
	sb.WriteString("Purchased Package:\n")
	for i, c := range cardsToAdd {
		sb.WriteString(fmt.Sprintf("[%d]: %s\n", i, c))
	}*/

	return cardsToAdd, nil
}