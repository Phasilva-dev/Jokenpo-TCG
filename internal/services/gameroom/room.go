//START OF FILE jokenpo/internal/services/gameroom/room.go
package gameroom

import (
	"bytes"
	"encoding/json"
	"fmt"
	"jokenpo/internal/game/card"
	"jokenpo/internal/game/deck"
	"log"
	"math/rand/v2"
	"net/http"
	"sync/atomic"
	"time"
)

// (Constantes de fase, sem mudanças)
const (
	phase_ROOM_START        = "room_start"
	phase_WAITING_FOR_PLAYS = "waiting_for_plays"
	phase_RESOLVING_ROUND   = "resolving_round"
	phase_GAME_OVER         = "game_over"
	phase_ROUND_START       = "round_start"
	initial_HAND_SIZE       = 5
)

type PlayerGameInfo struct {
	ID          string
	CallbackURL string
	GameDeck *deck.Deck
}

type GameRoom struct {
	ID          string
	players     map[string]*PlayerGameInfo
	rng         *rand.Rand
	incoming    chan interface{}
	quit        chan struct{}
	httpClient  *http.Client
	gameState   atomic.Value // Usamos atomic.Value para gameState ser thread-safe
	playedCards map[string]*card.Card
	roundTimer  *time.Timer
}

func NewGameRoom(id string, initialPlayerInfos []*InitialPlayerInfo, client *http.Client) (*GameRoom, error) {
	gr := &GameRoom{
		ID:          id,
		players:     make(map[string]*PlayerGameInfo),
		rng:         rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 1)),
		incoming:    make(chan interface{}),
		quit:        make(chan struct{}),
		httpClient:  client,
		playedCards: make(map[string]*card.Card),
	}
	gr.gameState.Store(phase_ROOM_START)

	for _, info := range initialPlayerInfos {
		
		gameDeck := deck.NewDeck()
		for _, cardKey := range info.Deck {
			c, err := card.GetCard(cardKey)
			if err != nil {
				return nil, fmt.Errorf("player %s has an invalid card in deck: %w", info.ID, err)
			}
			gameDeck.AddCardToZone("deck", c)
		}
		gr.players[info.ID] = &PlayerGameInfo{
			ID:          info.ID,
			CallbackURL: info.CallbackURL,
			GameDeck: gameDeck,
		}
	}
	return gr, nil
}

func (gr *GameRoom) Run() {
	log.Printf("[GameRoom %s] Goroutine starting for players: %v", gr.ID, gr.getPlayerIDs())
	defer func() {
		if gr.roundTimer != nil {
			gr.roundTimer.Stop()
		}
		gr.setGameState(phase_GAME_OVER) // Garante que IsFinished() retorne true
		log.Printf("[GameRoom %s] Goroutine stopped.", gr.ID)
	}()

	gr.startGame()

	for {
		select {
		case action := <-gr.incoming:
			switch act := action.(type) {
			case PlayCardAction:
				gr.HandlePlayCard(act.PlayerID, act.CardIndex)
			}
		case <-gr.roundTimer.C:
			if gr.getGameState() == phase_WAITING_FOR_PLAYS {
				gr.handleTimeout()
				if gr.getGameState() != phase_GAME_OVER {
					gr.resolveRound()
				}
			}
		case <-gr.quit:
			return
		}
	}
}

// --- MÉTODOS PARA INTERAÇÃO EXTERNA ---

// ForwardAction envia uma ação para o canal da sala de forma segura.
func (gr *GameRoom) ForwardAction(action interface{}) {
	if gr.IsFinished() {
		return // Não aceita mais ações se o jogo acabou.
	}
	// O envio pode bloquear se o canal estiver cheio, o que é o comportamento desejado.
	gr.incoming <- action
}

// IsFinished verifica se o jogo terminou. É seguro para ser chamado de outras goroutines.
func (gr *GameRoom) IsFinished() bool {
	return gr.getGameState() == phase_GAME_OVER
}

// --- Métodos de Estado Thread-Safe ---
func (gr *GameRoom) getGameState() string {
	return gr.gameState.Load().(string)
}

func (gr *GameRoom) setGameState(state string) {
	gr.gameState.Store(state)
}

// --- Funções de Callback (sem mudanças) ---
func (gr *GameRoom) broadcastEvent(eventType string, data interface{}) {
	for _, pInfo := range gr.players {
		gr.sendCallbackToPlayer(pInfo.ID, eventType, data)
	}
}
func (gr *GameRoom) sendCallbackToPlayer(playerID string, eventType string, data interface{}) {
	pInfo := gr.players[playerID]
	if pInfo == nil { return }
	eventPayload := map[string]interface{}{"eventType": eventType, "playerId": playerID, "roomId": gr.ID, "data": data}
	jsonData, _ := json.Marshal(eventPayload)
	go func(url string, payload []byte) {
		resp, err := gr.httpClient.Post(url, "application/json", bytes.NewBuffer(payload))
		if err != nil {
			log.Printf("[GameRoom %s] ERROR: Failed to send callback to %s: %v", gr.ID, url, err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 300 {
			log.Printf("[GameRoom %s] WARN: Callback to %s returned non-success status: %s", gr.ID, url, resp.Status)
		}
	}(pInfo.CallbackURL, jsonData)
}

// --- Funções Helper ---
func (gr *GameRoom) getPlayerIDs() []string {
	ids := make([]string, 0, len(gr.players))
	for id := range gr.players {
		ids = append(ids, id)
	}
	return ids
}

//END OF FILE jokenpo/internal/services/gameroom/room.go