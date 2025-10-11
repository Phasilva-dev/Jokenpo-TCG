package session

import (
	"encoding/json"
	"jokenpo/internal/session/message"
)

func handleLeaveQueue(h *GameHandler, session *PlayerSession, payload json.RawMessage) {


	if !checkQueueState(session) {
		session.Client.Send() <- message.CreateErrorResponse("You are not in queue")
		session.Client.Send() <- message.CreatePromptInputMessage()
		return
	}

		session.State = state_LOBBY
		//
		msg := message.CreateSuccessResponse(session.State, "Your request to leave the queue was received.", nil)
		session.Client.Send() <- msg

		h.matchmaker.LeaveQueue(session)
		session.Client.Send() <- message.CreatePromptInputMessage()

}



func (h *GameHandler) registerQueueHandlers() {
	// --- Ações de Matchmaking ---
	h.lobbyRouter["LEAVE_QUEUE"] = handleLeaveQueue
}

func checkQueueState(session *PlayerSession) bool {
	return session.State == state_IN_QUEUE
}
