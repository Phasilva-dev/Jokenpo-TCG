package session


type PlayCardAction struct {
	Session *PlayerSession
	CardIndex int
}
// ForwardPlayCardAction é um novo método específico para esta ação.
// O GameHandler usará este método.
func (gr *GameRoom) ForwardPlayCardAction(session *PlayerSession, cardIndex int) {
	action := PlayCardAction{
		Session:   session,
		CardIndex: cardIndex,
	}
	gr.incoming <- action
}