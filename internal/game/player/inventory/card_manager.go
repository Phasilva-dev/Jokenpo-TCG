package inventory

import (
	"jokenpo/internal/game/card"
	"jokenpo/internal/game/deck"
	"fmt"
)




// AddCardToDeck attempts to add a card from a player's collection to their game deck.
func (i *Inventory) AddCardToDeck(key string) (string, error) {
	cardToAdd, err := i.collection.GetCard(key)
	if err != nil {
		return "", err
	}

	currentDeck, err := i.gameDeck.GetCardsInZone(deck.DECK)
	if err != nil {
		return "", fmt.Errorf("internal error: could not access 'deck' zone: %w", err)
	}

	// Create a hypothetical state of the deck after the addition.
	hypotheticalDeck := append(currentDeck, cardToAdd)

	// Validate this new state using the orchestrator.
	if err := validateDeckState(hypotheticalDeck, i.collection); err != nil {
		return "", err // If validation fails, return the specific error.
	}

	// Validation passed, so we can safely execute the change.
	err = i.gameDeck.AddCardToZone(deck.DECK, cardToAdd)
	if err != nil {
		return "", err
	}
	string := fmt.Sprintf("This card %s is added to you deck", cardToAdd.String())
	return string, nil
}

// RemoveCardFromDeck removes a card from the deck at a specific index.
// Note: With the current rules, removing a card can't cause a violation,
// so validation is not strictly necessary. If a "minimum deck size" rule were added,
// this function would also need to use the validation helpers.
func (i *Inventory) RemoveCardFromDeck(index int) (string, error) {
	deckPile, err := i.gameDeck.GetZone("deck")
	if err != nil {
		return "",fmt.Errorf("internal error: could not access 'deck' zone: %w", err)
	}
	card, err := deckPile.RemoveCardByIndex(index)
	if err != nil {
		return "", err
	}
	string := fmt.Sprintf("This card %s is removed from you deck", card.String())
	return string, nil
}

// ReplaceCardInDeck safely replaces a card at a given index with a new one.
// It uses the validation orchestrator to ensure the operation is valid before executing it.
func (i *Inventory) ReplaceCardInDeck(indexToRemove int, keyOfCardToAdd string) (string, error) {
	cardToAdd, err := i.collection.GetCard(keyOfCardToAdd)
	if err != nil {
		return "", err
	}

	currentDeck, err := i.gameDeck.GetCardsInZone("deck")
	if err != nil {
		return "", fmt.Errorf("internal error: could not access 'deck' zone: %w", err)
	}

	if indexToRemove < 0 || indexToRemove >= len(currentDeck) {
		return "", fmt.Errorf("index %d is out of bounds for the current deck", indexToRemove)
	}

	if currentDeck[indexToRemove] == cardToAdd {
		return "", nil // No change needed, the operation is trivially successful.
	}

	// Create a hypothetical state of the deck after the replacement.
	hypotheticalDeck := make([]*card.Card, len(currentDeck))
	copy(hypotheticalDeck, currentDeck)
	hypotheticalDeck[indexToRemove] = cardToAdd

	// Validate this new state.
	if err := validateDeckState(hypotheticalDeck, i.collection); err != nil {
		return "", err
	}

	// Validation passed, so we can safely execute the change.
	deckPile, _ := i.gameDeck.GetZone("deck")
	cardToRemove, err := deckPile.RemoveCardByIndex(indexToRemove)
	if err != nil {
		// This should not happen since we already validated the index.
		return "", fmt.Errorf("internal error during removal: %w", err)
	}
	deckPile.AddCard(cardToAdd)

	string := fmt.Sprintf("The card [%s] is removed and replaced by [%s]", cardToRemove.String(), cardToAdd.String())

	return string, nil
}

// --- Rule 1: Deck Size Validation ---
// validateDeckSize checks if the number of cards in a deck exceeds the maximum limit.
func validateDeckSize(deck []*card.Card) error {
	const maxDeckSize = 12
	if len(deck) > maxDeckSize {
		return fmt.Errorf("deck size would be %d, which exceeds the limit of %d", len(deck), maxDeckSize)
	}
	return nil
}

// --- Rule 2: Deck Value Sum Validation ---
// validateDeckValueSum checks if the total value of all cards in a deck exceeds the maximum limit.
func validateDeckValueSum(deck []*card.Card) error {
	const maxValueSum = uint8(80)
	currentValueSum := uint8(0)
	for _, c := range deck {
		currentValueSum += c.Value()
	}
	if currentValueSum > maxValueSum {
		return fmt.Errorf("deck value sum would be %d, which exceeds the limit of %d", currentValueSum, maxValueSum)
	}
	return nil
}

// --- Rule 3: Card Copy Validation ---
// validateCardCopies checks if the number of copies of any card in the deck
// exceeds the number of copies the player owns in their collection.
func validateCardCopies(deck []*card.Card, collection *card.PlayerCollection) error {
	// An empty deck is always valid in terms of copies.
	if len(deck) == 0 {
		return nil
	}

	countsInDeck := make(map[*card.Card]int)
	for _, c := range deck {
		countsInDeck[c]++
	}

	for card, count := range countsInDeck {
		// This requires your *Card to have a Key() method
		instance, err := collection.GetInstance(card.Key())
		if err != nil {
			return fmt.Errorf("internal consistency error: card %s is in the deck but not in the player's collection", card.Key())
		}
		if count > int(instance.Count()) {
			return fmt.Errorf("deck would have %d copies of '%s', but player only owns %d", count, card.Key(), instance.Count())
		}
	}
	return nil
}

// --- Orchestrator ---
// validateDeckState runs all individual validation functions on a hypothetical deck state.
func validateDeckState(hypotheticalDeck []*card.Card, collection *card.PlayerCollection) error {
	if err := validateDeckSize(hypotheticalDeck); err != nil {
		return err
	}
	if err := validateDeckValueSum(hypotheticalDeck); err != nil {
		return err
	}
	if err := validateCardCopies(hypotheticalDeck, collection); err != nil {
		return err
	}
	return nil // All validations passed
}