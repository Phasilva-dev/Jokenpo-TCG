//START OF FILE jokenpo/internal/services/queue/service.go
package queue

import (
	"bytes"
	"encoding/json"
	"fmt"
	"jokenpo/internal/services/cluster"
	"log"
	"net/http"
	"time"
)

// ============================================================================
// Estruturas de Dados
// ============================================================================

// PlayerInfo agora contém o deck para passar para o GameRoomService.
type PlayerInfo struct {
	ID          string   `json:"playerId"`
	CallbackURL string   `json:"callbackUrl"`
	MatchCallbackURL string 
	Deck        []string `json:"deck"`
}

type TradeInfo struct {
	PlayerInfo
	OfferCard string `json:"offerCard"`
}

// --- DTOs para comunicação com outros serviços ---

type CreateRoomRequest struct {
	PlayerInfos []*PlayerInfo `json:"playerInfos"`
}
type CreateRoomResponse struct {
	RoomID      string `json:"roomId"`
	ServiceAddr string `json:"serviceAddr"`
}
type MatchCreatedPayload struct {
	PlayerIDs   []string `json:"playerIds"`
	RoomID      string   `json:"roomId"`
	ServiceAddr string   `json:"serviceAddr"`
}
type MatchFailedPayload struct {
	PlayerIDs []string `json:"playerIds"`
	Reason    string   `json:"reason"`
}

// ============================================================================
// Ator QueueMaster
// ============================================================================

type QueueMaster struct {
	matchQueue   []*PlayerInfo
	tradeQueue   []*TradeInfo
	requestCh    chan actorMessage
	httpClient   *http.Client
	serviceCache *cluster.ServiceCacheActor
}

// NewQueueMaster agora precisa do endereço do Consul para criar seu cache.
func NewQueueMaster(consulAddr string) *QueueMaster {
	return &QueueMaster{
		matchQueue:   make([]*PlayerInfo, 0),
		tradeQueue:   make([]*TradeInfo, 0),
		requestCh:    make(chan actorMessage),
		httpClient:   &http.Client{Timeout: 10 * time.Second},
		serviceCache: cluster.NewServiceCacheActor(30*time.Second, consulAddr),
	}
}

type actorMessage interface{ isActorMessage() }
type enqueueMatchRequest struct{ player *PlayerInfo }

func (enqueueMatchRequest) isActorMessage() {}
type dequeueMatchRequest struct{ playerID string }

func (dequeueMatchRequest) isActorMessage() {}
type enqueueTradeRequest struct{ trade *TradeInfo }

func (enqueueTradeRequest) isActorMessage() {}
type dequeueTradeRequest struct{ playerID string }

func (dequeueTradeRequest) isActorMessage() {}

func (m *QueueMaster) Run() {
	log.Println("[QueueMaster] Actor started. Waiting for players...")
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case msg := <-m.requestCh:
			switch req := msg.(type) {
			case enqueueMatchRequest:
				m.matchQueue = append(m.matchQueue, req.player)
				log.Printf("[QueueMaster] Player '%s' added to MATCH queue. Queue size: %d", req.player.ID, len(m.matchQueue))
			case dequeueMatchRequest:
				m.matchQueue = removePlayerFromMatchQueue(m.matchQueue, req.playerID)
			case enqueueTradeRequest:
				m.tradeQueue = append(m.tradeQueue, req.trade)
				log.Printf("[QueueMaster] Player '%s' added to TRADE queue with card '%s'. Queue size: %d", req.trade.ID, req.trade.OfferCard, len(m.tradeQueue))
			case dequeueTradeRequest:
				m.tradeQueue = removePlayerFromTradeQueue(m.tradeQueue, req.playerID)
			}
		case <-ticker.C:
			m.tryPairingMatches()
			m.tryPairingTrades()
		}
	}
}

func (m *QueueMaster) EnqueueMatch(player *PlayerInfo) { m.requestCh <- enqueueMatchRequest{player: player} }
func (m *QueueMaster) DequeueMatch(playerID string)   { m.requestCh <- dequeueMatchRequest{playerID: playerID} }
func (m *QueueMaster) EnqueueTrade(trade *TradeInfo)  { m.requestCh <- enqueueTradeRequest{trade: trade} }
func (m *QueueMaster) DequeueTrade(playerID string)   { m.requestCh <- dequeueTradeRequest{playerID: playerID} }

// ============================================================================
// Lógica de Pareamento e Orquestração
// ============================================================================

func (m *QueueMaster) tryPairingTrades() {
	if len(m.tradeQueue) < 2 { return }
	trade1 := m.tradeQueue[0]
	trade2 := m.tradeQueue[1]
	m.tradeQueue = m.tradeQueue[2:]
	log.Printf("[QueueMaster] TRADE FOUND! Pairing '%s' and '%s'.", trade1.ID, trade2.ID)
	payload1 := map[string]string{"playerId": trade1.ID, "cardSent": trade1.OfferCard, "cardReceived": trade2.OfferCard}
	go m.sendCallback(trade1.CallbackURL, payload1)
	payload2 := map[string]string{"playerId": trade2.ID, "cardSent": trade2.OfferCard, "cardReceived": trade1.OfferCard}
	go m.sendCallback(trade2.CallbackURL, payload2)
}

func (m *QueueMaster) tryPairingMatches() {
	if len(m.matchQueue) < 2 {
		return
	}
	player1 := m.matchQueue[0]
	player2 := m.matchQueue[1]
	m.matchQueue = m.matchQueue[2:]
	log.Printf("[QueueMaster] MATCH FOUND! Pairing '%s' and '%s'. Orchestrating room creation...", player1.ID, player2.ID)
	go m.orchestrateRoomCreation(player1, player2)
}

func (m *QueueMaster) orchestrateRoomCreation(p1, p2 *PlayerInfo) {
	// 1. Descobre um GameRoomService disponível.
	opts := cluster.DiscoveryOptions{Mode: cluster.ModeAnyHealthy}
	gameRoomServiceAddr := m.serviceCache.Discover("jokenpo-gameroom", opts)
	if gameRoomServiceAddr == "" {
		reason := "Could not create room: GameRoom service not found"
		log.Println("ERROR:", reason)
		m.notifyMatchFailed(p1, p2, reason)
		return
	}

	// 2. Envia a requisição para o GameRoomService criar a sala.
	createReq := CreateRoomRequest{PlayerInfos: []*PlayerInfo{p1, p2}}
	reqBody, _ := json.Marshal(createReq)
	createURL := fmt.Sprintf("http://%s/rooms", gameRoomServiceAddr)
	
	resp, err := m.httpClient.Post(createURL, "application/json", bytes.NewBuffer(reqBody))
	if err != nil || resp.StatusCode != http.StatusCreated {
		var status string
		if resp != nil {
			status = resp.Status
		}
		reason := fmt.Sprintf("Failed to create room in GameRoomService: %v, status: %s", err, status)
		log.Println("ERROR:", reason)
		m.notifyMatchFailed(p1, p2, reason)
		return
	}
	defer resp.Body.Close()

	// 3. Decodifica a resposta do GameRoomService.
	var roomResp CreateRoomResponse
	if err := json.NewDecoder(resp.Body).Decode(&roomResp); err != nil {
		reason := fmt.Sprintf("Failed to parse GameRoomService response: %v", err)
		log.Println("ERROR:", reason)
		m.notifyMatchFailed(p1, p2, reason)
		return
	}

	// 4. Notifica os servidores de sessão originais com as informações da sala.
	log.Printf("[QueueMaster] Room %s created at %s. Notifying session servers.", roomResp.RoomID, roomResp.ServiceAddr)
	matchCreatedPayload := MatchCreatedPayload{
		PlayerIDs:   []string{p1.ID, p2.ID},
		RoomID:      roomResp.RoomID,
		ServiceAddr: roomResp.ServiceAddr,
	}
	go m.sendCallback(p1.MatchCallbackURL, matchCreatedPayload)
	go m.sendCallback(p2.MatchCallbackURL, matchCreatedPayload)
}

// --- Funções Helper ---

func (m *QueueMaster) notifyMatchFailed(p1, p2 *PlayerInfo, reason string) {
	failPayload := MatchFailedPayload{
		PlayerIDs: []string{p1.ID, p2.ID},
		Reason:    reason,
	}
	go m.sendCallback(p1.MatchCallbackURL, failPayload)
	go m.sendCallback(p2.MatchCallbackURL, failPayload)
}

func removePlayerFromMatchQueue(queue []*PlayerInfo, playerID string) []*PlayerInfo {
	for i, p := range queue {
		if p.ID == playerID {
			log.Printf("[QueueMaster] Player '%s' removed from MATCH queue.", playerID)
			return append(queue[:i], queue[i+1:]...)
		}
	}
	return queue
}

func removePlayerFromTradeQueue(queue []*TradeInfo, playerID string) []*TradeInfo {
	for i, t := range queue {
		if t.ID == playerID {
			log.Printf("[QueueMaster] Player '%s' removed from TRADE queue.", playerID)
			return append(queue[:i], queue[i+1:]...)
		}
	}
	return queue
}

func (m *QueueMaster) sendCallback(callbackURL string, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("ERROR: Failed to marshal callback payload for URL %s: %v", callbackURL, err)
		return
	}
	resp, err := m.httpClient.Post(callbackURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Printf("ERROR: Failed to send callback to %s: %v", callbackURL, err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		log.Printf("WARN: Callback to %s returned non-success status: %s", callbackURL, resp.Status)
	} else {
		log.Printf("INFO: Successfully sent callback to %s", callbackURL)
	}
}
//END OF FILE jokenpo/internal/services/queue/service.go