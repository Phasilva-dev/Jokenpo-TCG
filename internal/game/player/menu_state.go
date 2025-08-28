package player

import (
	"fmt"
	"math/rand/v2"

	"jokenpo/internal/game/shop"
)

func (p *Player) SeeDeck() (string,error) {
	if p.state == MENU {
		string := p.inventory.GameDeck().String()
		return string, nil
	}
	return "",fmt.Errorf("")
}


func (p *Player) PurchasePackage(shop *shop.Shop, r *rand.Rand) (string, error){
	if p.state == MENU {
		return shop.PurchasePackage(r, p.inventory.Collection())
	}
	return "", fmt.Errorf("")
}

func (p *Player) SeeCollection() (string, error) {
	if p.state == MENU {
		string := p.inventory.Collection().String()
		return string, nil
	}
	return "", fmt.Errorf("")
}