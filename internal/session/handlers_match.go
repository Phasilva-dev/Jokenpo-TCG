//START OF FILE jokenpo/internal/session/handler_match.go
package session
/*
import (
	"encoding/json"
	"jokenpo/internal/session/message"
)

func handlePlayCard(h *GameHandler, session *PlayerSession, payload json.RawMessage) {
	if session.State != state_IN_MATCH || session.CurrentRoom == nil {
		session.Client.Send() <- message.CreateErrorResponse("You are not currently in a match.")
		return
	}

	// 2. Tradução e Validação do Payload: O JSON recebido é válido?
	// Definimos uma struct anônima para o formato esperado do payload.
	var req struct {
		// Usamos um ponteiro para int para poder verificar se o campo foi enviado.
		CardIndex *int `json:"cardIndex"`
	}

	if err := json.Unmarshal(payload, &req); err != nil || req.CardIndex == nil {
		// Se houver um erro ou o campo 'cardIndex' estiver faltando, o payload é inválido.
		session.Client.Send() <- message.CreateErrorResponse("Invalid payload: 'cardIndex' field is required and must be a number.")
		return
	}

	// 3. Encaminhamento para a GameRoom:
	// Se todas as validações passaram, temos dados limpos e seguros.
	// Agora, encaminhamos a ação para a GameRoom correta usando o método fortemente tipado.
	// A GameRoom não precisa se preocupar com JSON, apenas com a lógica do jogo.
	session.CurrentRoom.ForwardPlayCardAction(session, *req.CardIndex)
}

// registerMatchHandlers popula o roteador com os comandos disponíveis durante uma partida.
func (h *GameHandler) registerMatchHandlers() {

	h.matchRouter["PLAY_CARD"] = handlePlayCard

	// Se no futuro você adicionar mais ações de partida (ex: "USE_SKILL"),
}
*/

//END OF FILE jokenpo/internal/session/handler_match.go