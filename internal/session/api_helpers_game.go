//START OF FILE jokenpo/internal/session/api_helpers_game.go
package session

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

)

// ============================================================================
// DTOs para Comunicação com o GameRoomService
// ============================================================================

// PlayerInfoForRoom é o DTO enviado para criar uma sala.
type PlayerInfoForRoom struct {
	PlayerID    string   `json:"playerId"`
	CallbackURL string   `json:"callbackUrl"`
	Deck        []string `json:"deck"`
}

// CreateRoomRequest é o DTO que este Broker envia para o GameRoomService.
type CreateRoomRequest struct {
	PlayerInfos []*PlayerInfoForRoom `json:"playerInfos"`
}

// CreateRoomResponse é o DTO que esperamos receber do GameRoomService.
type CreateRoomResponse struct {
	RoomID      string `json:"roomId"`
	ServiceAddr string `json:"serviceAddr"`
}

// PlayCardRequest é o DTO para enviar a ação de jogar uma carta.
type PlayCardRequest struct {
	PlayerID  string `json:"playerId"`
	CardIndex int    `json:"cardIndex"`
}


// ============================================================================
// Helpers de API para o GameHandler
// ============================================================================


// forwardPlayCardAction é um helper para encaminhar a jogada de um jogador para o GameRoomService correto.
func (h *GameHandler) forwardPlayCardAction(session *PlayerSession, cardIndex int) error {
	if session.CurrentGame == nil {
		return fmt.Errorf("player is not in a game")
	}

	payload := PlayCardRequest{
		PlayerID:  session.ID,
		CardIndex: cardIndex,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to create play card payload: %w", err)
	}

	// Constrói a URL usando o endereço do serviço e o ID da sala armazenados na sessão.
	actionURL := fmt.Sprintf("http://%s/rooms/%s/play", session.CurrentGame.ServiceAddr, session.CurrentGame.RoomID)

	resp, err := h.httpClient.Post(actionURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to forward action to game room service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("game room service returned an error status: %s", resp.Status)
	}
	
	return nil
}



//END OF FILE jokenpo/internal/session/api_helpers_game.go