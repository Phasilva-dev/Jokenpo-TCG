package session
/*
import (
	"fmt"
	"jokenpo/internal/session/message"
)

// handleRoomCreationError lida com a falha na preparação de uma partida.
// Ele notifica ambos os jogadores, reverte o estado do jogador que estava pronto
// e coloca o jogador inocente de volta na fila de matchmaking.
func (h *GameHandler) handleRoomCreationError(failedPlayer, innocentPlayer *PlayerSession, err error) {
	// 1. Reverte o estado do jogador "inocente", caso ele já tenha sido preparado.
	innocentPlayer.Player.EndPlay() // Esta função deve ser segura para chamar mesmo se ele não estiver em jogo.

	// 2. Notifica o jogador que falhou sobre o motivo específico.
	failMsg := fmt.Sprintf("Could not start match: %v", err)
	failedPlayer.Client.Send() <- message.CreateErrorResponse(failMsg)
	failedPlayer.Player.EndPlay()
	failedPlayer.State = state_LOBBY
	failedPlayer.Client.Send() <- message.CreatePromptInputMessage()

	// 3. Notifica o jogador "inocente" que o oponente não estava pronto.
	innocentMsg := "Your opponent was not ready for the match. You have been placed back in the queue."
	innocentPlayer.Client.Send() <- message.CreateErrorResponse(innocentMsg)
	

	// 4. Coloca o jogador "inocente" de volta na fila de matchmaking.
	
	innocentPlayer.Player.EndPlay()
	innocentPlayer.Player.StartPlay()
	innocentPlayer.State = state_IN_MATCH_QUEUE
	h.matchmaker.EnqueuePlayer(innocentPlayer)
	innocentPlayer.Client.Send() <- message.CreatePromptInputMessage()

	fmt.Printf("Failed to create room. Player (%s) failed: %v. Player (%s) was re-queued.\n",
		failedPlayer.Client.Conn().RemoteAddr(), err, innocentPlayer.Client.Conn().RemoteAddr())
}
*/
