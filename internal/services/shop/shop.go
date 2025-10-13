package shop

import (
	"fmt"
	"jokenpo/internal/game/card"
	"math"
	"math/rand/v2"
	"time"
)

// --- MUDANÇA ---
// State representa o estado persistível do Shop.
// Os campos devem ser exportados (maiúsculos) para serem serializados em JSON.
type State struct {
	PackageCount uint64 `json:"package_count"`
}

type Shop struct {
	state State // Usamos a struct de estado em vez de um campo simples
	rng   *rand.Rand
}

func NewShop() *Shop {
	seed := uint64(time.Now().UnixNano())
	return &Shop{
		state: State{PackageCount: 0}, // Inicializa a struct de estado
		rng:   rand.New(rand.NewPCG(seed, 0)),
	}
}

// --- MUDANÇA ---
// GetState retorna uma cópia do estado atual.
func (s *Shop) GetState() State {
	return s.state
}

// SetState substitui completamente o estado do Shop.
// Usado pelo ator quando um nó se torna líder para restaurar o estado.
func (s *Shop) SetState(newState State) {
	s.state = newState
}

const maxPurchases = math.MaxUint64
const packageSize = 3

func (s *Shop) purchasePackage(quantity uint64) ([]*card.Card, error) {
	if quantity == 0 {
		return nil, fmt.Errorf("invalid quantity: must be greater than zero")
	}

	// --- MUDANÇA ---
	// Usa o campo dentro da struct de estado
	if s.state.PackageCount+quantity >= maxPurchases {
		return nil, fmt.Errorf("cannot process purchase: maximum purchase limit reached")
	}

	totalCards := int(quantity) * packageSize
	allCards := make([]*card.Card, 0, totalCards)

	for i := uint64(0); i < quantity; i++ {
		for j := 0; j < packageSize; j++ {
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

	// --- MUDANÇA ---
	// Atualiza o campo dentro da struct de estado
	s.state.PackageCount += quantity
	return allCards, nil
}