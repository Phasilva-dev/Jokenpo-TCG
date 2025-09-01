package session

import (
	"fmt"
	"jokenpo/internal/game/shop"
	"jokenpo/internal/network"
	"jokenpo/internal/session/message"
)

func (h *GameHandler) Matchmaker() *Matchmaker { return h.matchmaker }

func (h *GameHandler) Sessions() map[*network.Client]*PlayerSession { return h.sessions }

func (h *GameHandler) Rooms() map[string]*GameRoom { return h.rooms }

func (h *GameHandler) Shop() *shop.Shop { return h.shop }

func (h *GameHandler) CreateNewRoom(p1, p2 *PlayerSession) {
	// Tenta preparar o primeiro jogador.
	if err := p1.Player.StartPlay(); err != nil {
		// Se p1 falhar, p2 é o "inocente".
		h.handleRoomCreationError(p1, p2, err)
		return
	}

	// Tenta preparar o segundo jogador.
	if err := p2.Player.StartPlay(); err != nil {
		// Se p2 falhar, p1 é o "inocente".
		h.handleRoomCreationError(p2, p1, err)
		return
	}

	// --- SUCESSO ---
	// Se chegamos aqui, ambos os jogadores estão prontos.

	p1.State = state_IN_MATCH
	p2.State = state_IN_MATCH

	// TODO: Criar e gerenciar a GameRoom.

	NewGameRoom("",p1,p2, nil)
	
	msg := message.CreateSuccessResponse("Match found! The game is starting.", nil)
	p1.Client.Send() <- msg
	p2.Client.Send() <- msg

	fmt.Printf("Game room created successfully for %s and %s.\n", p1.Client.Conn().RemoteAddr(), p2.Client.Conn().RemoteAddr())
}