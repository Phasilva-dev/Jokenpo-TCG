//START OF FILE jokenpo/internal/session/handler.go
package session

import (
	"encoding/json"
	"fmt"
	"jokenpo/internal/network"
	"jokenpo/internal/services/cluster"
	"jokenpo/internal/session/message"
	"log"
	"net/http"
	"strings"
	"time"
	"jokenpo/internal/services/blockchain"
)

// CommandHandlerFunc define a assinatura para todas as nossas funções que lidam com comandos.
type CommandHandlerFunc func(h *GameHandler, session *PlayerSession, payload json.RawMessage)

type GameHandler struct {
	sessionsByClient   map[*network.Client]*PlayerSession
	sessionsByID       map[string]*PlayerSession
	httpClient         *http.Client
	serviceCache       *cluster.ServiceCacheActor
	advertisedHostname string
	lobbyRouter        map[string]CommandHandlerFunc
	matchRouter        map[string]CommandHandlerFunc
	matchQueueRouter   map[string]CommandHandlerFunc
	tradeQueueRouter   map[string]CommandHandlerFunc

	blockchain *blockchain.BlockchainClient
}

// NewGameHandler agora recebe o ConsulManager para criar o ServiceCacheActor.
func NewGameHandler(manager *cluster.ConsulManager, advertisedHostname string) (*GameHandler, error) {

	bcClient, err := blockchain.NewBlockchainClient()
    if err != nil {
        log.Printf("AVISO: Blockchain indisponível (%v). O jogo rodará sem auditoria.", err)
    }

	h := &GameHandler{
		sessionsByClient:   make(map[*network.Client]*PlayerSession),
		sessionsByID:       make(map[string]*PlayerSession),
		advertisedHostname: advertisedHostname,
		lobbyRouter:        make(map[string]CommandHandlerFunc),
		matchRouter:        make(map[string]CommandHandlerFunc),
		matchQueueRouter:   make(map[string]CommandHandlerFunc),
		tradeQueueRouter:   make(map[string]CommandHandlerFunc),

		 blockchain: bcClient,
	}

	h.httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}
	// O ServiceCacheActor agora é criado com o manager, garantindo resiliência.
	h.serviceCache = cluster.NewServiceCacheActor(10*time.Second, manager)

	h.registerLobbyHandlers()
	h.registerQueueHandlers()
	h.registerMatchHandlers()

	return h, nil
}

// --- Implementação da Interface network.EventHandler ---

func (h *GameHandler) OnConnect(c *network.Client) {
	session := NewPlayerSession(c)
	h.sessionsByClient[c] = session
	h.sessionsByID[session.ID] = session

	log.Printf("Session created for %s. Total sessions: %d", c.Conn().RemoteAddr(), len(h.sessionsByClient))

	const initialPacksToOpen = 4
	// --- MUDANÇA: Passamos session.ID para o Shop fazer a mintagem na blockchain
	initialCardKeys, err := h.purchasePacksFromShop(session.ID, initialPacksToOpen)
	if err != nil {
		log.Printf("CRITICAL: Failed to grant initial packs to player %s: %v", c.Conn().RemoteAddr(), err)
		welcomeMsg := "Welcome to the Jokenpo Game!\n\n" +
			"Unfortunately, we could not grant you your initial card packs at this time as our shop is unavailable. " +
			"Please try the 'purchase' command later."
		message.SendSuccessAndPrompt(c, state_LOBBY, "Connection successful!", welcomeMsg)
		return
	}

	for _, key := range initialCardKeys {
		if err := session.Player.AddCardToCollection(key, 1); err != nil {
			log.Printf("CRITICAL ERROR: Failed to add purchased card '%s' to collection for player %s: %v", key, c.Conn().RemoteAddr(), err)
			var sb strings.Builder
			sb.WriteString("Welcome to the Jokenpo Game!\n")
			sb.WriteString(fmt.Sprintf("You received %d card packs, but a critical error occurred while adding them to your collection:\n\n", initialPacksToOpen))
			sb.WriteString(fmt.Sprintf("Error: %v\n\n", err))
			sb.WriteString("Please contact support.")
			message.SendSuccessAndPrompt(c, state_LOBBY, "Connection successful, but an error occurred!", sb.String())
			return
		}
	}

	deckBuildMessage := fmt.Sprintf("All %d initial cards have been added to your collection and starting deck.", len(initialCardKeys))

	for i, key := range initialCardKeys {
		if _, err := session.Player.AddCardToDeck(key); err != nil {
			deckBuildMessage = fmt.Sprintf(
				"All cards were added to your collection, but an error occurred while building your starting deck after adding %d cards.\nReason: %v",
				i, err,
			)
			log.Printf("INFO: Could not add card '%s' to initial deck for player %s: %v", key, c.Conn().RemoteAddr(), err)
			break
		}
	}

	var sb strings.Builder
	sb.WriteString("Welcome to the Jokenpo Game!\n")
	sb.WriteString(fmt.Sprintf("As a bonus, you received %d card packs, revealing the following cards:\n\n", initialPacksToOpen))
	for i, key := range initialCardKeys {
		msg := fmt.Sprintf("[%d] - %s \n", i, key)
		sb.WriteString(msg)
	}
	sb.WriteString("\n")
	sb.WriteString(deckBuildMessage)

	message.SendSuccessAndPrompt(
		c,
		state_LOBBY,
		"Connection successful! Welcome!",
		sb.String(),
	)
}

func (h *GameHandler) OnDisconnect(c *network.Client) {
	session, ok := h.sessionsByClient[c]
	if !ok {
		return
	}

	if session.State == state_IN_MATCH_QUEUE {
		log.Printf("Player %s disconnected while in match queue. Notifying Queue Service.", session.ID)
		h.leaveMatchQueue(session)
	} else if session.State == state_IN_TRADE_QUEUE {
		log.Printf("Player %s disconnected while in trade queue. Notifying Queue Service.", session.ID)
		h.leaveTradeQueue(session)
	}

	delete(h.sessionsByClient, c)
	delete(h.sessionsByID, session.ID)

	log.Printf("Session %s for %s removed. Total sessions: %d", session.ID, c.Conn().RemoteAddr(), len(h.sessionsByClient))
}

func (h *GameHandler) OnMessage(c *network.Client, msg network.Message) {
	session, ok := h.sessionsByClient[c]
	if !ok {
		return
	}

	var router map[string]CommandHandlerFunc
	switch session.State {
	case state_LOBBY:
		router = h.lobbyRouter
	case state_IN_MATCH:
		router = h.matchRouter
	case state_IN_MATCH_QUEUE:
		router = h.matchQueueRouter
	case state_IN_TRADE_QUEUE:
		router = h.tradeQueueRouter
	default:
		message.SendErrorAndPrompt(c, "Invalid player state: %s", session.State)
		return
	}

	handler, found := router[msg.Type]
	if !found {
		message.SendErrorAndPrompt(c, "Unknown or invalid command for your current state ('%s'): %s", session.State, msg.Type)
		return
	}

	handler(h, session, msg.Payload)
}
//END OF FILE jokenpo/internal/session/handler.go