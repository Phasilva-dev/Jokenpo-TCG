// START OF FILE jokenpo/internal/services/gameroom/room.go
package gameroom

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"jokenpo/internal/game/card"
	"jokenpo/internal/game/deck"
	"jokenpo/internal/services/blockchain" // Importar
	"log"
	"math/rand/v2"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"
)

// ... (Constantes iguais) ...
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
	start       chan struct{}
	httpClient  *http.Client
	gameState   atomic.Value
	playedCards map[string]*card.Card
	roundTimer  *time.Timer
    blockchain  *blockchain.BlockchainClient // Novo campo
}

// NewGameRoom atualizado
func NewGameRoom(id string, initialPlayerInfos []*InitialPlayerInfo, client *http.Client, bc *blockchain.BlockchainClient) (*GameRoom, error) {
	gr := &GameRoom{
		ID:          id,
		players:     make(map[string]*PlayerGameInfo),
		rng:         rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 1)),
		incoming:    make(chan interface{}),
		quit:        make(chan struct{}),
		start:       make(chan struct{}),
		httpClient:  client,
		playedCards: make(map[string]*card.Card),
        blockchain:  bc,
	}
	log.Printf("GameRoom de ID %s foi criado",gr.ID)
	gr.gameState.Store(phase_ROOM_START)

	for i, info := range initialPlayerInfos {
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
			GameDeck:    gameDeck,
		}
		log.Printf("[DEBUG] Player %d, ID: (%s) deck size: %d",i , info.ID, gameDeck.DeckSize())
	}
	return gr, nil
}

func (gr *GameRoom) StartGame() {
	close(gr.start)
}

func (gr *GameRoom) Run() {
	log.Printf("[GameRoom %s] Goroutine starting, WAITING FOR START SIGNAL.", gr.ID)
	<-gr.start
	log.Printf("[GameRoom %s] Start signal received, commencing game.", gr.ID)
	log.Printf("[GameRoom %s] Goroutine starting for players: %v", gr.ID, gr.getPlayerIDs())
	defer func() {
		if gr.roundTimer != nil {
			gr.roundTimer.Stop()
		}
		gr.setGameState(phase_GAME_OVER)
		log.Printf("[GameRoom %s] Goroutine stopped.", gr.ID)
	}()

	log.Printf("[DEBUG] SALA COM %s ID ESTA RODANDO", gr.ID)
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

func (gr *GameRoom) ForwardAction(action interface{}) {
	if gr.IsFinished() {
		log.Printf("[GameRoom %s] WARN: Action received after game over. Ignoring.", gr.ID)
		return
	}
	select {
	case gr.incoming <- action:
	default:
		log.Printf("[GameRoom %s] WARN: Incoming action channel is busy. Action discarded (likely a late play).", gr.ID)
	}
}

func (gr *GameRoom) IsFinished() bool {
	return gr.getGameState() == phase_GAME_OVER
}

func (gr *GameRoom) getGameState() string {
	return gr.gameState.Load().(string)
}

func (gr *GameRoom) setGameState(state string) {
	gr.gameState.Store(state)
}

func (gr *GameRoom) broadcastEvent(eventType string, data interface{}) {
	log.Printf("[GameRoom %s] Broadcasting event '%s' to %d players.", gr.ID, eventType, len(gr.players))
	for _, pInfo := range gr.players {
		go func(player *PlayerGameInfo) {
			if err := gr.sendCallbackToPlayer(player.ID, eventType, data); err != nil {
				log.Printf("[GameRoom %s] ERROR: Failed to send event '%s' to player %s after all retries: %v", gr.ID, eventType, player.ID, err)
			}
		}(pInfo)
	}
}

func (gr *GameRoom) sendCallbackToPlayer(playerID string, eventType string, data interface{}) error {
	pInfo, ok := gr.players[playerID]
	if !ok {
		return fmt.Errorf("player %s not found in room", playerID)
	}

	if pInfo.CallbackURL == "" {
		return fmt.Errorf("player %s has an empty callback URL", playerID)
	}
	log.Printf("O CALLBACK DO PLAYER %s É %s", pInfo.ID, pInfo.CallbackURL)

	if _, err := url.ParseRequestURI(pInfo.CallbackURL); err != nil {
		return fmt.Errorf("invalid callback URL for player %s: %w", playerID, err)
	}

	eventPayload := map[string]interface{}{
		"eventType": eventType,
		"playerId":  playerID,
		"roomId":    gr.ID,
		"data":      data,
	}
	jsonData, err := json.Marshal(eventPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload for event %s: %w", eventType, err)
	}

	var lastErr error
	backoff := 200 * time.Millisecond

	for attempt := 1; attempt <= 3; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, pInfo.CallbackURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("failed to create HTTP request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		log.Printf("[GameRoom %s] Sending event '%s' to player %s at %s (Attempt %d/3)...", gr.ID, eventType, playerID, pInfo.CallbackURL, attempt)
		resp, err := gr.httpClient.Do(req)

		if err != nil {
			lastErr = err
			log.Printf("[GameRoom %s] WARN: Attempt %d to send '%s' to player %s failed: %v", gr.ID, attempt, eventType, playerID, err)
		} else {
			resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				log.Printf("[GameRoom %s] SUCCESS: Event '%s' delivered to player %s.", gr.ID, eventType, playerID)
				return nil
			}
			lastErr = fmt.Errorf("received non-success status code: %s", resp.Status)
			log.Printf("[GameRoom %s] WARN: Attempt %d to send '%s' to player %s received status: %s", gr.ID, attempt, eventType, playerID, resp.Status)
		}

		if attempt < 3 {
			time.Sleep(backoff)
			backoff *= 2
		}
	}

	return fmt.Errorf("failed to send callback after all retries: %w", lastErr)
}

func (gr *GameRoom) getPlayerIDs() []string {
	ids := make([]string, 0, len(gr.players))
	for id := range gr.players {
		ids = append(ids, id)
	}
	return ids
}
//END OF FILE jokenpo/internal/services/gameroom/room.go