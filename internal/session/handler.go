package session

import (
	"encoding/json"
	"fmt"
	"jokenpo/internal/game/shop"
	"jokenpo/internal/network"
)

// GameHandler implementa a interface network.EventHandler.
// Ele gerencia o estado do jogo e confia no network.Hub para serializar
// o acesso aos seus métodos, tornando desnecessário o uso de mutexes ou canais internos.
type GameHandler struct {
	sessions   map[*network.Client]*PlayerSession
	matchmaker *Matchmaker
	rooms      map[string]*GameRoom
	shop       *shop.Shop
}

// NewGameHandler cria um novo GameHandler.
func NewGameHandler() *GameHandler {
	return &GameHandler{
		sessions:   make(map[*network.Client]*PlayerSession),
		matchmaker: NewMatchmaker(),
		rooms:      make(map[string]*GameRoom),
		shop:       shop.NewShop(),
	}
}

// --- Implementação da Interface network.EventHandler ---

// OnConnect é chamado pela goroutine do network.Hub. É seguro modificar o estado aqui.
func (h *GameHandler) OnConnect(c *network.Client) {
	session := NewPlayerSession(c)
	h.sessions[c] = session
	fmt.Printf("Sessão criada para %s. Total de sessões: %d\n", c.Conn().RemoteAddr(), len(h.sessions))
	
	// Envia mensagem de boas-vindas
	welcomeMsg := createSuccessResponse("Bem-vindo ao servidor!", nil)
	c.Send() <- welcomeMsg
}

// OnDisconnect é chamado pela goroutine do network.Hub. É seguro modificar o estado aqui.
func (h *GameHandler) OnDisconnect(c *network.Client) {
	if _, ok := h.sessions[c]; ok {
		// Lógica futura: se o jogador estava em uma sala, notifique o oponente.
		delete(h.sessions, c)
		fmt.Printf("Sessão removida para %s. Total de sessões: %d\n", c.Conn().RemoteAddr(), len(h.sessions))
	}
}

// OnMessage é chamado pela goroutine do network.Hub. É seguro modificar o estado aqui.
func (h *GameHandler) OnMessage(c *network.Client, msg network.Message) {
	session, ok := h.sessions[c]
	if !ok {
		// Cliente enviou mensagem mas não tem sessão. Ignorar.
		return
	}

	// Aqui, você roteia o comando baseado no estado da sessão.
	switch session.State {
	case StateLobby:
		h.handleLobbyCommand(session, msg)
	case StateInMatch:
		h.handleMatchCommand(session, msg)
	default:
		// Estado desconhecido, talvez envie um erro.
	}
}