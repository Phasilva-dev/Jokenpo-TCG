//START OF FILE jokenpo/internal/session/handler_queue.go
package session

import (
	"encoding/json"
	"jokenpo/internal/session/message"
)

// handleLeaveQueue processa o comando do jogador para sair de qualquer fila em que ele esteja.
func handleLeaveQueue(h *GameHandler, session *PlayerSession, payload json.RawMessage) {
	// 1. Verifica se o jogador está em um estado de fila.
	if !checkQueueState(session) {
		message.SendErrorAndPrompt(session.Client, "You are not in any queue.")
		return
	}

	var err error
	currentState := session.State

	// 2. Determina de qual fila sair e chama o helper de API correspondente.
	if currentState == state_IN_MATCH_QUEUE {
		err = h.leaveMatchQueue(session)
	} else if currentState == state_IN_TRADE_QUEUE {
		// ROLLBACK da carta: Antes de sair da fila, a carta que o jogador
		// ofereceu precisa ser devolvida à sua coleção.
		// Precisamos saber qual carta era. Isso deve ser armazenado na sessão.
		// (Vamos adicionar isso no próximo passo).
		// Por enquanto, vamos assumir que recuperamos a carta e a devolvemos.
		// session.Player.Inventory().Collection().AddCard(session.OfferedCardForTrade, 1)

		err = h.leaveTradeQueue(session)
	}
	
	// 3. Lida com o resultado da chamada de API.
	if err != nil {
		// Se o QueueService retornou um erro, informa o jogador e NÃO muda seu estado.
		message.SendErrorAndPrompt(session.Client, "Failed to leave queue: %v", err)
		return
	}

	// 4. Apenas se a chamada de API foi bem-sucedida, muda o estado do jogador de volta para o lobby.
	session.State = state_LOBBY
	
	message.SendSuccessAndPrompt(
		session.Client,
		session.State,
		"You have successfully left the queue and returned to the lobby.",
		nil,
	)
}

// registerQueueHandlers registra os comandos que são válidos enquanto o jogador está em uma fila.
func (h *GameHandler) registerQueueHandlers() {
	// O único comando válido enquanto está em uma fila é sair dela.
	h.matchQueueRouter["LEAVE_QUEUE"] = handleLeaveQueue
	
	// Registra o mesmo handler para o estado da fila de troca também.
	h.tradeQueueRouter["LEAVE_QUEUE"] = handleLeaveQueue 
}

// checkQueueState verifica se o estado da sessão é um estado de fila válido.
func checkQueueState(session *PlayerSession) bool {
	return session.State == state_IN_MATCH_QUEUE || session.State == state_IN_TRADE_QUEUE
}

//END OF FILE jokenpo/internal/session/handler_queue.go