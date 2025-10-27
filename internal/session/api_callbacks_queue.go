//START OF FILE jokenpo/internal/session/api_callbacks_queue.go
package session

import (
	"encoding/json"
	"fmt"
	"io"
	"jokenpo/internal/session/message"
	"log"
	"net/http"
)

// ============================================================================
// Callback da Fila de Partida
// ============================================================================

// MatchCreatedPayload é o DTO de SUCESSO que o jokenpo-session espera receber do QueueService.
type MatchCreatedPayload struct {
	PlayerIDs   []string `json:"playerIds"`
	RoomID      string   `json:"roomId"`
	ServiceAddr string   `json:"serviceAddr"`
}

// MatchFailedPayload é o DTO de FALHA que o jokenpo-session espera receber do QueueService.
type MatchFailedPayload struct {
	PlayerIDs []string `json:"playerIds"`
	Reason    string   `json:"reason"`
}

// CallbackMatchFound é o handler HTTP para o endpoint de callback (ex: /match-found).
// Ele é chamado pelo QueueService para notificar o resultado do matchmaking.
func (h *GameHandler) CallbackMatchFound(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Lê o corpo da requisição uma vez para poder tentar múltiplos Unmarshals.
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Cannot read request body", http.StatusBadRequest)
		return
	}

	// Tenta decodificar como um payload de sucesso.
	var successPayload MatchCreatedPayload
	if err := json.Unmarshal(bodyBytes, &successPayload); err == nil && successPayload.RoomID != "" {
		h.handleMatchSuccess(w, &successPayload)
		return
	}

	// Se não for sucesso, tenta decodificar como falha.
	var failPayload MatchFailedPayload
	if err := json.Unmarshal(bodyBytes, &failPayload); err == nil && failPayload.Reason != "" {
		h.handleMatchFailure(w, &failPayload)
		return
	}
	
	log.Printf("WARN: Received unknown payload on /match-found endpoint: %s", string(bodyBytes))
	http.Error(w, "Invalid or unknown payload", http.StatusBadRequest)
}

// handleMatchSuccess é chamado quando o QueueService informa que a sala foi criada com sucesso.
func (h *GameHandler) handleMatchSuccess(w http.ResponseWriter, payload *MatchCreatedPayload) {
	log.Printf("[Callback] Match creation successful for room %s at %s.", payload.RoomID, payload.ServiceAddr)

	// Cria a struct com as informações da partida remota.
	gameInfo := &CurrentGameInfo{
		RoomID:      payload.RoomID,
		ServiceAddr: payload.ServiceAddr,
	}

	// Itera sobre os jogadores do par. Atualiza o estado daquele(s) jogador(es)
	// que estiver(em) nesta instância do jokenpo-session.
	for _, playerID := range payload.PlayerIDs {
		session := h.findSessionByID(playerID)
		if session != nil {
			// --- A ATUALIZAÇÃO CRUCIAL ACONTECE AQUI ---
			session.State = state_IN_MATCH
			session.CurrentGame = gameInfo
			
			// Notifica o cliente (via WebSocket) que ele está em uma partida.
			// O GameRoomService enviará as mensagens de início de jogo (compra de cartas, etc.).
			message.SendSuccess(session.Client, session.State, "Match found! Entering game room...", gameInfo)
		}
	}
	w.WriteHeader(http.StatusOK)
}

// handleMatchFailure é chamado quando o QueueService informa que a criação da sala falhou.
func (h *GameHandler) handleMatchFailure(w http.ResponseWriter, payload *MatchFailedPayload) {
	log.Printf("[Callback] Match creation failed: %s", payload.Reason)

	// Itera sobre os jogadores e devolve ao lobby aquele(s) que estiver(em) nesta instância.
	for _, playerID := range payload.PlayerIDs {
		session := h.findSessionByID(playerID)
		if session != nil {
			session.State = state_LOBBY
			message.SendErrorAndPrompt(session.Client, "Failed to create match: %s. You have been returned to the lobby.", payload.Reason)
		}
	}
	w.WriteHeader(http.StatusOK)
}

// ============================================================================
// Callback da Fila de Troca (sem mudanças)
// ============================================================================

type TradeFoundPayload struct {
	PlayerID     string `json:"playerId"`
	CardSent     string `json:"cardSent"`
	CardReceived string `json:"cardReceived"`
}

func (h *GameHandler) CallbackTradeFound(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { /* ... */ return }
	var payload TradeFoundPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil { /* ... */ return }

	session := h.findSessionByID(payload.PlayerID)
	if session == nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	log.Printf("[Callback] Trade completed for player %s.", session.ID)
	if err := session.Player.Inventory().Collection().AddCard(payload.CardReceived, 1); err != nil {
		log.Printf("CRITICAL: Failed to add received card '%s' for player %s: %v", payload.CardReceived, session.ID, err)
		message.SendErrorAndPrompt(session.Client, "A trade was found, but an internal error occurred.")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	session.State = state_LOBBY
	successMsg := fmt.Sprintf("Wonder Trade successful! You sent '%s' and received '%s'.", payload.CardSent, payload.CardReceived)
	message.SendSuccessAndPrompt(session.Client, session.State, "Trade Completed!", successMsg)

	w.WriteHeader(http.StatusOK)
}




//END OF FILE jokenpo/internal/session/api_callbacks_queue.go