//START OF FILE jokenpo/internal/services/queue/service.go
package queue

import (
	"bytes"
	"encoding/json"
	"fmt"
	"jokenpo/internal/services/blockchain"
	"jokenpo/internal/services/cluster"
	"log"
	"net/http"
	"time"
)

// ... (Structs e NewQueueMaster permanecem iguais) ...
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
type QueueMaster struct {
	matchQueue   []*PlayerInfo
	tradeQueue   []*TradeInfo
	requestCh    chan actorMessage
	httpClient   *http.Client
	serviceCache *cluster.ServiceCacheActor
    blockchain   *blockchain.BlockchainClient
}
func NewQueueMaster(manager *cluster.ConsulManager) *QueueMaster {
    var bcClient *blockchain.BlockchainClient
    var contractAddr string
    log.Println("QUEUE: Aguardando endereço do contrato no Consul...")
    client := manager.GetClient()
    for i := 0; i < 60; i++ {
        if client == nil { client = manager.GetClient() }
        if client != nil {
            pair, _, err := client.KV().Get("jokenpo/config/contract_address", nil)
            if err == nil && pair != nil {
                contractAddr = string(pair.Value)
                break
            }
        }
        time.Sleep(2 * time.Second)
    }
    if contractAddr != "" {
        var err error
        bcClient, _, err = blockchain.InitBlockchain(contractAddr)
        if err != nil { log.Printf("QUEUE AVISO: %v", err) } else { log.Printf("QUEUE: Conectado blockchain %s", contractAddr) }
    } else { log.Println("QUEUE AVISO: Timeout blockchain.") }

	return &QueueMaster{
		matchQueue:   make([]*PlayerInfo, 0),
		tradeQueue:   make([]*TradeInfo, 0),
		requestCh:    make(chan actorMessage),
		httpClient:   &http.Client{Timeout: 10 * time.Second},
		serviceCache: cluster.NewServiceCacheActor(30*time.Second, manager),
        blockchain:   bcClient,
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
	log.Println("[QueueMaster] Actor started.")
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case msg := <-m.requestCh:
			switch req := msg.(type) {
			case enqueueMatchRequest:
				m.matchQueue = append(m.matchQueue, req.player)
				log.Printf("[QM] +MatchQueue: %s", req.player.ID)
			case dequeueMatchRequest:
				m.matchQueue = removePlayerFromMatchQueue(m.matchQueue, req.playerID)
			case enqueueTradeRequest:
				m.tradeQueue = append(m.tradeQueue, req.trade)
				log.Printf("[QM] +TradeQueue: %s offers %s", req.trade.ID, req.trade.OfferCard)
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


// --- MUDANÇA PRINCIPAL AQUI ---

func (m *QueueMaster) tryPairingTrades() {
	if len(m.tradeQueue) < 2 {
		return
	}
	trade1 := m.tradeQueue[0]
	trade2 := m.tradeQueue[1]
	m.tradeQueue = m.tradeQueue[2:]
	log.Printf("[QueueMaster] TRADE MATCH: %s <-> %s", trade1.ID, trade2.ID)

    // --- LÓGICA DE REGISTRO NA BLOCKCHAIN (Troca de Tokens Específicos) ---
    if m.blockchain != nil {
        go func() {
            // 1. Encontrar o Token UUID real da carta do Jogador 1
            token1, err := m.blockchain.FindTokenForCard(trade1.ID, trade1.OfferCard)
            if err != nil {
                log.Printf("QUEUE ERRO: Não foi possível encontrar o token blockchain para %s do player %s: %v", trade1.OfferCard, trade1.ID, err)
                return // Aborta registro se não achar
            }

            // 2. Encontrar o Token UUID real da carta do Jogador 2
            token2, err := m.blockchain.FindTokenForCard(trade2.ID, trade2.OfferCard)
            if err != nil {
                log.Printf("QUEUE ERRO: Não foi possível encontrar o token blockchain para %s do player %s: %v", trade2.OfferCard, trade2.ID, err)
                return
            }

            log.Printf("QUEUE BLOCKCHAIN: Trocando tokens [%s] <-> [%s]", token1, token2)

            // 3. Executa a Troca 1: A -> B (Envia Token 1)
            if err := m.blockchain.LogTrade(trade1.ID, trade2.ID, token1); err != nil {
                log.Printf("QUEUE ERRO: Falha TX A->B: %v", err)
            } else {
                 log.Printf("QUEUE SUCESSO: %s transferido para %s", token1, trade2.ID)
            }

            // 4. Executa a Troca 2: B -> A (Envia Token 2)
            if err := m.blockchain.LogTrade(trade2.ID, trade1.ID, token2); err != nil {
                log.Printf("QUEUE ERRO: Falha TX B->A: %v", err)
            } else {
                log.Printf("QUEUE SUCESSO: %s transferido para %s", token2, trade1.ID)
            }
        }()
    }

	// Callbacks HTTP para o Jogo (Mantém a lógica de inventário funcionando)
	payload1 := map[string]string{
		"playerId":     trade1.ID,
		"cardSent":     trade1.OfferCard,
		"cardReceived": trade2.OfferCard,
		"partnerId":    trade2.ID, 
	}
	go m.sendCallback(trade1.CallbackURL, payload1)

	payload2 := map[string]string{
		"playerId":     trade2.ID,
		"cardSent":     trade2.OfferCard,
		"cardReceived": trade1.OfferCard,
		"partnerId":    trade1.ID,
	}
	go m.sendCallback(trade2.CallbackURL, payload2)
}

func (m *QueueMaster) tryPairingMatches() {
	if len(m.matchQueue) < 2 { return }
	p1 := m.matchQueue[0]
	p2 := m.matchQueue[1]
	m.matchQueue = m.matchQueue[2:]
	log.Printf("[QueueMaster] MATCH FOUND! %s vs %s", p1.ID, p2.ID)
	go m.orchestrateRoomCreation(p1, p2)
}

func (m *QueueMaster) orchestrateRoomCreation(p1, p2 *PlayerInfo) {
	opts := cluster.DiscoveryOptions{Mode: cluster.ModeAnyHealthy}
	addr := m.serviceCache.Discover("jokenpo-gameroom", opts)
	if addr == "" {
		m.notifyMatchFailed(p1, p2, "GameRoom service not found")
		return
	}
	createReq := CreateRoomRequest{PlayerInfos: []*PlayerInfo{p1, p2}}
	reqBody, _ := json.Marshal(createReq)
	resp, err := m.httpClient.Post(fmt.Sprintf("http://%s/rooms", addr), "application/json", bytes.NewBuffer(reqBody))
	if err != nil || resp.StatusCode != http.StatusCreated {
		m.notifyMatchFailed(p1, p2, "Failed to create room")
		return
	}
	defer resp.Body.Close()
	var roomResp CreateRoomResponse
	json.NewDecoder(resp.Body).Decode(&roomResp)
	payload := MatchCreatedPayload{ PlayerIDs: []string{p1.ID, p2.ID}, RoomID: roomResp.RoomID, ServiceAddr: roomResp.ServiceAddr }
	go m.sendCallback(p1.MatchCallbackURL, payload)
	go m.sendCallback(p2.MatchCallbackURL, payload)
}
func (m *QueueMaster) notifyMatchFailed(p1, p2 *PlayerInfo, reason string) {
	pl := MatchFailedPayload{ PlayerIDs: []string{p1.ID, p2.ID}, Reason: reason }
	go m.sendCallback(p1.MatchCallbackURL, pl)
	go m.sendCallback(p2.MatchCallbackURL, pl)
}
func removePlayerFromMatchQueue(q []*PlayerInfo, id string) []*PlayerInfo {
	for i, p := range q { if p.ID == id { return append(q[:i], q[i+1:]...) } }
	return q
}
func removePlayerFromTradeQueue(q []*TradeInfo, id string) []*TradeInfo {
	for i, t := range q { if t.ID == id { return append(q[:i], q[i+1:]...) } }
	return q
}
func (m *QueueMaster) sendCallback(url string, payload interface{}) {
	if url == "" { return }
	data, _ := json.Marshal(payload)
	m.httpClient.Post(url, "application/json", bytes.NewBuffer(data))
}
//END OF FILE jokenpo/internal/services/queue/service.go