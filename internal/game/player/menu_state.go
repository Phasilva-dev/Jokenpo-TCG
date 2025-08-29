package player

import (
	"fmt"
	"math/rand/v2"

	"jokenpo/internal/game/shop"
)

func (p *Player) SeeDeck() (string,error) {
	if p.state != MENU {
		return "", fmt.Errorf("error: Player must be in MENU state to see the deck")
	}
	str := p.inventory.GameDeck().String()
	return str, nil
	
}


func (p *Player) PurchasePackage(shop *shop.Shop, r *rand.Rand) (string, error){
	if p.state != MENU {
		return "", fmt.Errorf("error: Player must be in MENU state to purchase a package")
	}
	return shop.PurchasePackage(r, p.inventory.Collection())
}

func (p *Player) SeeCollection() (string, error) {
	if p.state != MENU {
		return "", fmt.Errorf("error: Player must be in MENU state to see the collection")
	}
	str := p.inventory.Collection().String()
	return str, nil
}

func (p *Player) AddCardToDeck(key string) (string, error) {
	if p.state != MENU {
		return "", fmt.Errorf("")
	}
	return p.inventory.AddCardToDeck(key)
}

func (p *Player) RemoveCardFromDeck(index int) (string, error) {
	if p.state != MENU {
		return "", fmt.Errorf("")
	}
	return  p.inventory.RemoveCardFromDeck(index)
}

func (p *Player) ReplaceCardInDeck(indexToRemove int, keyOfCardToAdd string) (string, error) {
	if p.state != MENU {
		return "", fmt.Errorf("")
	}
	return p.inventory.ReplaceCardInDeck(indexToRemove, keyOfCardToAdd)
}
