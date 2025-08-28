package session

import (
	"jokenpo/internal/game/shop"
	"jokenpo/internal/network"
)

// Manager implementa a interface network.EventHandler.
// Ele gerencia todas as sessões de jogadores e a lógica de gameplay.
type Manager struct {
	// Mapeia um cliente de rede a uma sessão de jogador.
	players    map[*network.Client]*NetworkPlayer
	matchmaker *Matchmaker
	shop       *shop.Shop // Você adicionaria a loja aqui também
}