//START OF FILE jokenpo/internal/session/api_callbacks.go
package session

import (
	"encoding/json"
	"fmt"
	"jokenpo/internal/session/message"
	"log"
	"net/http"
)

// ============================================================================
// Callback da Fila de Partida
// ============================================================================

// MatchFoundPayload é o DTO que esperamos receber do QueueService quando uma partida é encontrada.
type MatchFoundPayload struct {
	PlayerIDs []string `json:"playerIds"`
}

// CallbackMatchFound é o handler HTTP para o endpoint de callback /match-found.
// Ele é chamado pelo QueueService.
func (h *GameHandler) CallbackMatchFound(w http.ResponseWriter, r *http.Request) {
	// Validação básica da requisição
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload MatchFoundPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	if len(payload.PlayerIDs) != 2 {
		http.Error(w, "Payload must contain exactly two player IDs", http.StatusBadRequest)
		return
	}

	// Encontra as sessões dos jogadores na memória desta instância do jokenpo-session.
	p1Session := h.findSessionByID(payload.PlayerIDs[0])
	p2Session := h.findSessionByID(payload.PlayerIDs[1])

	// Lógica de Criação da Sala:
	// Em uma arquitetura escalável, é possível que p1 e p2 estejam em instâncias diferentes
	// do jokenpo-session. A criação de sala se torna mais complexa (precisaria de um GameRoomService).
	// Para este PBL, vamos assumir a simplificação de que a sala só é criada se ambos os jogadores
	// estiverem, por sorte, na mesma instância que recebeu o callback.
	if p1Session != nil && p2Session != nil {
		log.Printf("[Callback] Both players (%s, %s) found on this instance. Creating game room...", p1Session.ID, p2Session.ID)
		// Você precisará reintroduzir a lógica de CreateNewRoom no seu GameHandler.
		// h.CreateNewRoom(p1Session, p2Session)
	} else {
		// Loga quais jogadores foram encontrados aqui. O outro jokenpo-session receberá o mesmo callback.
		if p1Session != nil {
			log.Printf("[Callback] Match found for player %s, but opponent is on another instance.", p1Session.ID)
		}
		if p2Session != nil {
			log.Printf("[Callback] Match found for player %s, but opponent is on another instance.", p2Session.ID)
		}
	}

	// Responde 200 OK para o QueueService saber que a notificação foi recebida.
	w.WriteHeader(http.StatusOK)
}

// ============================================================================
// Callback da Fila de Troca
// ============================================================================

// TradeFoundPayload é o DTO que esperamos receber do QueueService quando uma troca é encontrada.
type TradeFoundPayload struct {
	PlayerID     string `json:"playerId"`
	CardSent     string `json:"cardSent"`
	CardReceived string `json:"cardReceived"`
}

// CallbackTradeFound é o handler HTTP para o endpoint de callback /trade-found.
func (h *GameHandler) CallbackTradeFound(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload TradeFoundPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Encontra a sessão do jogador que está nesta instância.
	session := h.findSessionByID(payload.PlayerID)
	if session == nil {
		// O callback era para um jogador em outra instância. Ignoramos silenciosamente.
		w.WriteHeader(http.StatusOK)
		return
	}

	log.Printf("[Callback] Trade completed for player %s. Received '%s', sent '%s'.", session.ID, payload.CardReceived, payload.CardSent)

	// Adiciona a nova carta à coleção do jogador.
	if err := session.Player.Inventory().Collection().AddCard(payload.CardReceived, 1); err != nil {
		log.Printf("CRITICAL: Failed to add received card '%s' to collection for player %s: %v", payload.CardReceived, session.ID, err)
		// Informa o jogador sobre o erro interno.
		message.SendErrorAndPrompt(session.Client, "A trade was found, but an internal error occurred while receiving your new card.")
		// Em um sistema real, isso exigiria uma lógica de compensação (devolver a carta original).
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Devolve o jogador ao lobby.
	session.State = state_LOBBY

	// Envia a mensagem de sucesso ao jogador pela conexão WebSocket.
	successMsg := fmt.Sprintf("Your Wonder Trade was successful!\n\nYou sent: %s\nYou received: %s", payload.CardSent, payload.CardReceived)
	message.SendSuccessAndPrompt(session.Client, session.State, "Trade Completed!", successMsg)

	w.WriteHeader(http.StatusOK)
}

//END OF FILE jokenpo/internal/session/api_callbacks.go