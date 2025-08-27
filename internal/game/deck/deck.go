package deck

import (
	"fmt"
	"jokenpo/internal/game/card"
)

type Deck struct {
	zones map[string]*pileOfCards
}

// Inicializa todas as zonas vazias
func NewDeck() *Deck {
	return &Deck{
		zones: map[string]*pileOfCards{
			"deck":  new(pileOfCards),
			"hand":  new(pileOfCards),
			"win":   new(pileOfCards),
			"out":   new(pileOfCards),
			"play":  new(pileOfCards),
		},
	}
}

// GetZone returns a pointer to the pile of cards of a specific zone.
// This is used for direct manipulation of a zone, such as removing a card by index
// or shuffling.
func (d *Deck) GetZone(zoneName string) (*pileOfCards, error) {
	// Look up the zone in the map.
	pile, ok := d.zones[zoneName]
	if !ok {
		// If the zone doesn't exist, return an error.
		return nil, fmt.Errorf("zone '%s' not found", zoneName)
	}

	// If the zone exists, return the pointer to the pile and no error.
	return pile, nil
}

// RemoveCardByIndex removes a card from the pile at a specific index.
// It modifies the original slice.
func (p *pileOfCards) RemoveCardByIndex(index int) error {
	// 1. Validate the index to prevent a program crash (panic).
	// Checks if the index is negative or greater than or equal to the pile's length.
	if index < 0 || index >= len(*p) {
		return fmt.Errorf("index %d is out of bounds for a pile of size %d", index, len(*p))
	}

	// 2. Perform the removal using a standard Go slice trick.
	// It works by creating a new slice that combines everything *before* the index
	// with everything *after* the index, effectively cutting out the element at the index.
	*p = append((*p)[:index], (*p)[index+1:]...)

	// 3. Return nil to indicate that the operation was successful.
	return nil
}

// Move uma carta do topo do deck para a mão
func (d *Deck) DrawToHand() error {
	deck := d.zones["deck"]
	hand := d.zones["hand"]

	card, err := deck.DrawTop()
	if err != nil {
		return err
	}
	hand.AddCard(card)
	return nil
}

// Move uma carta da mão para a zona de play (index da mão)
func (d *Deck) PlayCardFromHand(index int) error {
	hand := d.zones["hand"]
	play := d.zones["play"]

	card, err := hand.GetCard(index)
	if err != nil {
		return err
	}

	if err := hand.RemoveCard(card); err != nil {
		return err
	}
	play.AddCard(card)
	return nil
}

func (d *Deck) ResolvePlay(won bool) error {
	play := d.zones["play"]
	var target *pileOfCards

	if won {
		target = d.zones["win"]
	} else {
		target = d.zones["out"]
	}

	card, err := play.DrawTop() //Como é uma pilha de uma carta só, faz sentido isso.
	if err != nil {
		return err
	}

	target.AddCard(card)
	return nil
}

func (d *Deck) WinCondition() bool {
	win := d.zones["win"]

	if win == nil || len(*win) == 0 {
		return false
	}

	// Contadores para cores e tipos
	colorCount := map[string]int{}
	typeCount := map[string]int{}

	for _, c := range *win {
		colorCount[c.Color()]++
		typeCount[c.Typo()]++
	}

	// Verifica cores: 3 de qualquer cor
	for _, count := range colorCount {
		if count >= 3 {
			return true
		}
	}

	// Verifica tipos: 1 rock+1 paper+1 scissor ou 3 de qualquer tipo
	if typeCount["rock"] >= 1 && typeCount["paper"] >= 1 && typeCount["scissor"] >= 1 {
		return true
	}
	for _, count := range typeCount {
		if count >= 3 {
			return true
		}
	}

	return false
}

// AddCardToZone adiciona uma única carta a uma zona específica.
// Isso é ideal para popular o baralho no início do jogo.
// Retorna um erro se a zona não existir.
func (d *Deck) AddCardToZone(zoneName string, c *card.Card) error {
	// 1. Encontra a pilha de cartas da zona de destino.
	pile, ok := d.zones[zoneName]
	if !ok {
		return fmt.Errorf("zone '%s' does not exist", zoneName)
	}

	// 2. Usa o método AddCard da própria pilha para adicionar a carta.
	// Isso mantém seu código organizado e reutiliza a lógica que você já escreveu.
	pile.AddCard(c)

	return nil
}

// Move todas as cartas de todas as zonas (exceto "deck") de volta para o deck
func (d *Deck) ResetToDeck() {
	deck := d.zones["deck"]

	for zoneName, pile := range d.zones {
		if zoneName == "deck" {
			continue
		}
		// adiciona todas as cartas de volta ao deck
		*deck = append(*deck, *pile...)
		// esvazia a zona atual
		*pile = (*pile)[:0]
	}
}

// GetCardsInZone retorna uma cópia do slice de ponteiros de cartas de uma zona específica.
// Ideal para inspecionar o conteúdo de uma zona sem modificar o ponteiro da pilha original.
// Retorna um erro se a zona não for encontrada.
func (d *Deck) GetCardsInZone(zoneName string) ([]*card.Card, error) {
	// 1. Encontra a pilha de cartas da zona.
	pile, ok := d.zones[zoneName]
	if !ok {
		return nil, fmt.Errorf("zone '%s' not found", zoneName)
	}

	// 2. Dereferencia o ponteiro para obter o slice.
	// `pile` é do tipo *pileOfCards, que é um *[]*card.Card.
	// `*pile` é do tipo []*card.Card, que é o que você quer retornar.
	return *pile, nil
}

func (d *Deck) PrintZone(zoneName string) error {
	pile, ok := d.zones[zoneName]
	if !ok {
		return fmt.Errorf("error: this zone '%s' not exist", zoneName)
	}

	fmt.Printf("--- Zona: %s ---\n", zoneName)
	if len(*pile) == 0 {
		fmt.Println("(Vazia)")
	} else {
		for i, card := range *pile {
			// Usamos o método String() da carta que definimos acima.
			fmt.Printf("[%d]: %s\n", i, card)
		}
	}
	fmt.Println("--------------------")
	return nil
}

// PrintAllZones exibe o conteúdo de todas as zonas do deck.
func (d *Deck) PrintAllZones() {
	fmt.Println("\n================== ESTADO ATUAL DO JOGO ==================")
	// Definimos uma ordem para a impressão ser sempre consistente
	zoneOrder := []string{"deck", "hand", "play", "win", "out"}
	for _, zoneName := range zoneOrder {
		d.PrintZone(zoneName)
	}
	fmt.Println("==========================================================")
}