package session

type GameRoom struct {
	players map[*NetworkPlayer]bool
	actions chan GameAction
	unregister chan *NetworkPlayer
}