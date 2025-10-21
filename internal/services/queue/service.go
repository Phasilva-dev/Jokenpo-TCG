//START OF FILE jokenpo/internal/services/queue/service.go
package queue

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// ============================================================================
// Estruturas de Dados
// ============================================================================

// PlayerInfo é a representação genérica de um jogador em uma fila.
type PlayerInfo struct {
	ID          string `json:"playerId"`
	CallbackURL string `json:"callbackUrl"`
}

// TradeInfo armazena os dados de um jogador na fila de trocas às cegas.
type TradeInfo struct {
	PlayerInfo
	OfferCard string `json:"offerCard"` // A carta que o jogador está oferecendo.
}

// ============================================================================
// Mensagens do Ator
// ============================================================================

// actorMessage é a interface que todas as mensagens para o QueueMaster devem implementar.
type actorMessage interface {
	isActorMessage()
}

// --- Mensagens da Fila de Partida ---
type enqueueMatchRequest struct{ player *PlayerInfo }
func (enqueueMatchRequest) isActorMessage() {}

type dequeueMatchRequest struct{ playerID string }
func (dequeueMatchRequest) isActorMessage() {}

// --- Mensagens da Fila de Troca ---
type enqueueTradeRequest struct{ trade *TradeInfo }
func (enqueueTradeRequest) isActorMessage() {}

type dequeueTradeRequest struct{ playerID string }
func (dequeueTradeRequest) isActorMessage() {}


// ============================================================================
// O Ator QueueMaster
// ============================================================================

// QueueMaster é o ator que gerencia as filas de partida e troca.
type QueueMaster struct {
	matchQueue []*PlayerInfo
	tradeQueue []*TradeInfo

	requestCh  chan actorMessage
	httpClient *http.Client
}

// NewQueueMaster cria uma nova instância do QueueMaster.
func NewQueueMaster() *QueueMaster {
	return &QueueMaster{
		matchQueue: make([]*PlayerInfo, 0),
		tradeQueue: make([]*TradeInfo, 0),
		requestCh:  make(chan actorMessage),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Run inicia o loop principal do ator em sua própria goroutine.
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

// --- APIs Públicas para Interagir com o Ator ---

func (m *QueueMaster) EnqueueMatch(player *PlayerInfo) {
	m.requestCh <- enqueueMatchRequest{player: player}
}

func (m *QueueMaster) DequeueMatch(playerID string) {
	m.requestCh <- dequeueMatchRequest{playerID: playerID}
}

func (m *QueueMaster) EnqueueTrade(trade *TradeInfo) {
	m.requestCh <- enqueueTradeRequest{trade: trade}
}

func (m *QueueMaster) DequeueTrade(playerID string) {
	m.requestCh <- dequeueTradeRequest{playerID: playerID}
}


// ============================================================================
// Lógica Interna e Helpers
// ============================================================================

// --- Lógica de Pareamento ---

func (m *QueueMaster) tryPairingMatches() {
	if len(m.matchQueue) < 2 {
		return
	}
	player1 := m.matchQueue[0]
	player2 := m.matchQueue[1]
	m.matchQueue = m.matchQueue[2:]

	log.Printf("[QueueMaster] MATCH FOUND! Pairing '%s' and '%s'.", player1.ID, player2.ID)


	//Adicionar a Logica de criação de sala
	payload := map[string][]string{"playerIds": {player1.ID, player2.ID}}
	go m.sendCallback(player1.CallbackURL, payload)
	go m.sendCallback(player2.CallbackURL, payload)
}

// tryPairingTrades agora implementa a lógica de "troca às cegas".
func (m *QueueMaster) tryPairingTrades() {
	if len(m.tradeQueue) < 2 {
		return
	}
	// Pega os dois primeiros jogadores da fila de troca.
	trade1 := m.tradeQueue[0]
	trade2 := m.tradeQueue[1]
	m.tradeQueue = m.tradeQueue[2:]

	log.Printf("[QueueMaster] TRADE FOUND! Pairing '%s' and '%s'. They will swap '%s' and '%s'.", 
		trade1.ID, trade2.ID, trade1.OfferCard, trade2.OfferCard)

	// Prepara os callbacks. Cada jogador precisa saber o que ele deu e o que recebeu.
	// Payload para o Jogador 1:
	payload1 := map[string]string{
		"playerId":     trade1.ID,
		"cardSent":     trade1.OfferCard,
		"cardReceived": trade2.OfferCard,
	}
	go m.sendCallback(trade1.CallbackURL, payload1)

	// Payload para o Jogador 2:
	payload2 := map[string]string{
		"playerId":     trade2.ID,
		"cardSent":     trade2.OfferCard,
		"cardReceived": trade1.OfferCard,
	}
	go m.sendCallback(trade2.CallbackURL, payload2)
}


// --- Funções Helper ---

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