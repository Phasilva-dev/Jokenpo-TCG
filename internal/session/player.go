package session

import (
	"jokenpo/internal/game/player"
	"jokenpo/internal/network"
	"time"
)

// Constantes de estado da sessão para evitar erros de digitação.
const (
	STATE_LOBBY = "lobby"  // Jogador está online, no menu, pode usar o chat, etc.
	STATE_IN_MATCH = "in-match" // Jogador está em uma partida ativa.
)

// PlayerSession representa um jogador único e conectado ao servidor.
type PlayerSession struct {
	Client *network.Client
	Player *player.Player

	State  string // Usará as constantes StateLobby ou StateInMatch.
	UserID int
}

// NewPlayerSession cria e inicializa uma nova sessão de jogador.
func NewPlayerSession(client *network.Client) *PlayerSession {
	seed := uint64(time.Now().UnixNano())

	return &PlayerSession{
		Client: client,
		Player: player.NewPlayer(seed),
		State:  STATE_LOBBY, // Todo jogador começa no lobby.
		// UserID: 0, // Será definido após o login.
	}
}