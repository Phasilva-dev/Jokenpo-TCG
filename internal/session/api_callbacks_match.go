// START OF FILE jokenpo/internal/session/api_callbacks_match.go
package session

import (
	"encoding/json"
	"jokenpo/internal/session/message"
	"log"
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
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
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
	
	log.Printf("[Callback] Received game event '%s' for player %s.", event.EventType, event.PlayerID)

	// --- LÓGICA DE CORREÇÃO FINAL ---

	// PADRONIZA A MENSAGEM PARA O CLIENTE
	// A mensagem para o cliente será o tipo de evento (ex: "GAME_START").
	messageToClient := event.EventType
	// Os dados para o cliente serão os dados do evento.
	dataToClient := event.Data

	if event.EventType == "GAME_OVER" {
		session.State = state_LOBBY
		session.CurrentGame = nil
		
		// Para GAME_OVER, a mensagem principal é mais clara.
		messageToClient = "The game has ended."
		
		// Envia a mensagem de sucesso com o NOVO estado e o prompt.
		message.SendSuccessAndPrompt(session.Client, session.State, messageToClient, dataToClient)

	} else {
		// Para todos os outros eventos, envie um RESPONSE_SUCCESS padrão
		// que o cliente consiga interpretar.
		// O `message` pode ser o próprio tipo de evento para depuração.
		
		// A função SendSuccess já cria a mensagem do tipo RESPONSE_SUCCESS.
		// O que faltava era padronizar o que vai dentro do 'message' e 'data' do payload.
		message.SendSuccess(session.Client, session.State, messageToClient, dataToClient)
		message.SendPromptInput(session.Client)
	}

	w.WriteHeader(http.StatusOK)
}

//END OF FILE jokenpo/internal/session/api_callbacks_match.go