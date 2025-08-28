package session

import (
	"jokenpo/internal/game/player"
	"jokenpo/internal/network"
)

type NetworkPlayer struct {
	netClient *network.Client
	
	player *player.Player

	state string

	room *GameRoom
}

