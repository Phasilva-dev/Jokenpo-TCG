package deck

import (
	"fmt"
	"jokenpo/internal/game/card"
	"math/rand/v2"
	"strings"
)

const DECK = "deck"
const HAND = "hand"
const WIN = "win"
const OUT = "out"
const PLAY = "play"

var zoneOrder = []string{DECK, HAND, PLAY, WIN, OUT}


type Deck struct {
	zones map[string]*card.Pile
}

// Inicializa todas as zonas vazias
func NewDeck() *Deck {
	return &Deck{
		zones: map[string]*card.Pile{
			DECK:  new(card.Pile),
			HAND:  new(card.Pile),
			WIN:   new(card.Pile),
			OUT:   new(card.Pile),
			PLAY:  new(card.Pile),
		},
	}
}

// GetZone returns a pointer to the pile of cards of a specific zone.
// This is used for direct manipulation of a zone, such as removing a card by index
// or shuffling.
func (d *Deck) GetZone(zoneName string) (*card.Pile, error) {
	// Look up the zone in the map.
	pile, ok := d.zones[zoneName]
	if !ok {
		// If the zone doesn't exist, return an error.
		return nil, fmt.Errorf("zone '%s' not found", zoneName)
	}

	// If the zone exists, return the pointer to the pile and no error.
	return pile, nil
}



// Move uma carta do topo do deck para a mão
func (d *Deck) DrawToHand() (*card.Card,error) {
	deck := d.zones["deck"]
	hand := d.zones["hand"]

	card, err := deck.DrawTop()
	if err != nil {
		return nil, err
	}
	hand.AddCard(card)
	return card, nil
}

// Move uma carta da mão para a zona de play (index da mão)
func (d *Deck) PlayCardFromHand(index int) (*card.Card, error) {
	hand := d.zones["hand"]
	play := d.zones["play"]

	card, err := hand.GetCard(index)
	if err != nil {
		return nil, err
	}

	if err := hand.RemoveCard(card); err != nil {
		return nil, err
	}
	play.AddCard(card)
	return card, nil
}

func (d *Deck) PlayRandomCardFromHand(rng *rand.Rand) (*card.Card, error) {
	hand := d.zones["hand"]
	play := d.zones["play"]

	n := rng.IntN(hand.Size())
	card, err := hand.GetCard(n)
	if err != nil {
		return nil, err
	}

	if err := hand.RemoveCard(card); err != nil {
		return nil, err
	}
	play.AddCard(card)
	return card, nil
}

func (d *Deck) ResolvePlay(won bool) (string, error) {
	play := d.zones["play"]
	var target *card.Pile
	var string string

	if won {
		target = d.zones["win"]
		string = "you win this round"
	} else {
		target = d.zones["out"]
		string = "you lose this round"
	}

	card, err := play.DrawTop() //Como é uma pilha de uma carta só, faz sentido isso.
	if err != nil {
		return "", err
	}

	target.AddCard(card)
	return string, nil
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
		if zoneName == "deck" || pile.Size() == 0 {
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
	// `pile` é do tipo *card.Pile, que é um *[]*card.Card.
	// `*pile` é do tipo []*card.Card, que é o que você quer retornar.
	return *pile, nil
}

func (d *Deck) String() string {
	if d == nil || len(d.zones) == 0 {
		return "(Empty Deck)"
	}

	var sb strings.Builder
	sb.WriteString("\n================== DECK ==================\n")
	
	cards, _ :=d.GetCardsInZone(DECK)
	var deckPower int
	for i := 0; i < len(cards); i++ {
		deckPower += int(cards[i].Value())
	}
	deckPowerString := fmt.Sprintf("\nDeck Power = %d\n", deckPower)

	sb.WriteString(deckPowerString)


	for _, zoneName := range zoneOrder {
		pile, ok := d.zones[zoneName]
		if !ok || len(*pile) == 0 {
			// pula zonas vazias
			continue
		}

		// usa ZoneString para cada zona
		zoneStr, err := d.ZoneString(zoneName)
		if err != nil {
			// se ocorrer algum erro, apenas ignora a zona
			continue
		}
		sb.WriteString(zoneStr + "\n")
	}

	sb.WriteString("=========================================\n")
	return sb.String()
}

func (d *Deck) ZoneString(zoneName string) (string, error) {
	pile, ok := d.zones[zoneName]
	if !ok {
		return "", fmt.Errorf("error: this zone '%s' does not exist", zoneName)
	}

	var sb strings.Builder
	//sb.WriteString(fmt.Sprintf("--- %s ---\n", zoneName))

	if len(*pile) == 0 {
		sb.WriteString("(Vazia)\n")
	} else {
		for i, card := range *pile {
			// chama automaticamente Card.String()
			sb.WriteString(fmt.Sprintf("[%d]: %s\n", i, card))
		}
	}

	sb.WriteString("--------------------\n")
	return sb.String(), nil
}