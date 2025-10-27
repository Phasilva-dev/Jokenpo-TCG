//START OF FILE jokenpo/internal/session/api_helpers_game.go
package session

import (
	"bytes"
	"encoding/json"
	"fmt"
	"jokenpo/internal/services/cluster"
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

// createGameRoom é um helper que encapsula a chamada para criar uma nova sala no GameRoomService.
func (h *GameHandler) createGameRoom(p1Session, p2Session *PlayerSession) (*CreateRoomResponse, error) {
	// 1. Descobre qualquer instância saudável do GameRoomService.
	opts := cluster.DiscoveryOptions{Mode: cluster.ModeAnyHealthy}
	gameRoomServiceAddr := h.serviceCache.Discover("jokenpo-gameroom", opts)
	if gameRoomServiceAddr == "" {
		return nil, fmt.Errorf("the game room service is currently unavailable")
	}

	// --- CORREÇÃO AQUI ---
	// 2. Prepara o payload para ambos os jogadores.
	// Usa ToJSON() para obter o deck como []byte e depois Unmarshal para converter em []string.
	p1DeckJSON, err := p1Session.Player.Inventory().GameDeck().ToJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize p1 deck to JSON: %w", err)
	}
	var p1DeckKeys []string
	if err := json.Unmarshal(p1DeckJSON, &p1DeckKeys); err != nil {
		return nil, fmt.Errorf("failed to unmarshal p1 deck keys: %w", err)
	}

	p2DeckJSON, err := p2Session.Player.Inventory().GameDeck().ToJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize p2 deck to JSON: %w", err)
	}
	var p2DeckKeys []string
	if err := json.Unmarshal(p2DeckJSON, &p2DeckKeys); err != nil {
		return nil, fmt.Errorf("failed to unmarshal p2 deck keys: %w", err)
	}

	p1Info := &PlayerInfoForRoom{
		PlayerID:    p1Session.ID,
		CallbackURL: h.buildCallbackURL(p1Session, "/game-event"),
		Deck:        p1DeckKeys,
	}

	p2Info := &PlayerInfoForRoom{
		PlayerID:    p2Session.ID,
		CallbackURL: h.buildCallbackURL(p2Session, "/game-event"),
		Deck:        p2DeckKeys,
	}
	
	payload := CreateRoomRequest{
		PlayerInfos: []*PlayerInfoForRoom{p1Info, p2Info},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create room request payload: %w", err)
	}
	// --- FIM DA CORREÇÃO ---


	// 3. Faz a chamada POST para o GameRoomService.
	createURL := fmt.Sprintf("http://%s/rooms", gameRoomServiceAddr)
	resp, err := h.httpClient.Post(createURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to contact game room service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("game room service returned an error status: %s", resp.Status)
	}

	// 4. Decodifica a resposta.
	var createResp CreateRoomResponse
	if err := json.NewDecoder(resp.Body).Decode(&createResp); err != nil {
		return nil, fmt.Errorf("failed to parse create room response: %w", err)
	}
	
	return &createResp, nil
}


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