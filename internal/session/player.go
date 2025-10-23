//START OF FILE jokenpo/internal/session/player.go
package session

import (
	"jokenpo/internal/game/player"
	"jokenpo/internal/network"

	"github.com/google/uuid"
)

// Constantes de estado da sessão para evitar erros de digitação.
const (
	state_LOBBY = "lobby"  // Jogador está online, no menu, pode usar o chat, etc.
	state_IN_MATCH = "in-match" // Jogador está em uma partida ativa.
	state_IN_MATCH_QUEUE = "in-match-queue"
	state_IN_TRADE_QUEUE = "in-trade-queue"
)

// PlayerSession representa um jogador único e conectado ao servidor.
type PlayerSession struct {
	ID string
	Client *network.Client
	Player *player.Player

	State  string // Usará as constantes StateLobby ou StateInMatch.
	CurrentGame *CurrentGameInfo
}

// NewPlayerSession cria e inicializa uma nova sessão de jogador.
func NewPlayerSession(client *network.Client) *PlayerSession {

	return &PlayerSession{
		ID: uuid.NewString(),
		Client: client,
		Player: player.NewPlayer(),
		State:  state_LOBBY, // Todo jogador começa no lobby.
		CurrentGame: nil,
	}
}

// CurrentGameInfo armazena os detalhes da partida ativa de um jogador.
type CurrentGameInfo struct {
	RoomID     string `json:"roomId"`     // O UUID da sala de jogo
	ServiceAddr string `json:"serviceAddr"` // O endereço de rede (host:port) do GameRoomService onde a sala está.
}

//END OF FILE jokenpo/internal/session/player.go