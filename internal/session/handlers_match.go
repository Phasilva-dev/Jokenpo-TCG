//START OF FILE jokenpo/internal/session/handler_match.go
package session

import (
	"encoding/json"
	"jokenpo/internal/session/message"
)

// handlePlayCard processa o comando 'PLAY_CARD' de um jogador.
func handlePlayCard(h *GameHandler, session *PlayerSession, payload json.RawMessage) {
	// 1. Validação de Estado: O jogador está realmente em uma partida?
	if session.State != state_IN_MATCH || session.CurrentGame == nil {
		message.SendErrorAndPrompt(session.Client, "You are not currently in a match.")
		return
	}

	// 2. Decodifica o payload para obter o índice da carta.
	var req struct {
		CardIndex *int `json:"cardIndex"`
	}
	if err := json.Unmarshal(payload, &req); err != nil || req.CardIndex == nil {
		message.SendErrorAndPrompt(session.Client, "Invalid payload: 'cardIndex' field is required and must be a number.")
		return
	}

	// 3. Encaminha a ação para o GameRoomService USANDO O HELPER.
	// É aqui que a mágica acontece: você chama a função que já está pronta em api_helpers.go
	err := h.forwardPlayCardAction(session, *req.CardIndex)
	if err != nil {
		// Se a chamada de rede falhar, informa o jogador.
		message.SendErrorAndPrompt(session.Client, "Failed to send your action to the game server: %v", err)
		return
	}

	// Sucesso! A ação foi encaminhada. Não precisamos fazer mais nada aqui,
	// pois a confirmação virá via callback do GameRoomService.
}

func (h *GameHandler) registerMatchHandlers() {
	if h.matchRouter == nil {
		h.matchRouter = make(map[string]CommandHandlerFunc)
	}
	h.matchRouter["PLAY_CARD"] = handlePlayCard
}
//END OF FILE jokenpo/internal/session/handler_match.go