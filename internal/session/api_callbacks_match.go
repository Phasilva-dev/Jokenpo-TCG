package session

import (
	"encoding/json"
	"jokenpo/internal/session/message"
	"net/http"
)

// ============================================================================
// Callback Genérico de Eventos de Jogo (sem mudanças)
// ============================================================================

type GameEventPayload struct {
	EventType string          `json:"eventType"`
	PlayerID  string          `json:"playerId"`
	RoomID    string          `json:"roomId"`
	Data      json.RawMessage `json:"data"`
}

func (h *GameHandler) CallbackGameEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { /* ... */ return }
	
	var event GameEventPayload
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "Invalid event payload", http.StatusBadRequest)
		return
	}
	
	session := h.findSessionByID(event.PlayerID)
	if session == nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	message.SendSuccess(session.Client, session.State, event.EventType, event.Data)

	if event.EventType == "GAME_OVER" {
		session.State = state_LOBBY
		session.CurrentGame = nil
		message.SendPromptInput(session.Client)
	}

	w.WriteHeader(http.StatusOK)
}