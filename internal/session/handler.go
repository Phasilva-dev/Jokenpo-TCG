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

func NewGameHandler(manager *cluster.ConsulManager, advertisedHostname string) (*GameHandler, error) {
    var bcClient *blockchain.BlockchainClient
    var contractAddr string

    // --- LÓGICA DE ESPERA PELA BLOCKCHAIN (POLLING NO CONSUL) ---
    log.Println("SESSION: Aguardando endereço do contrato no Consul...")
    client := manager.GetClient()
    
    // Tenta por 120 segundos
    for i := 0; i < 60; i++ {
        if client == nil { client = manager.GetClient() }
        if client != nil {
            pair, _, err := client.KV().Get("jokenpo/config/contract_address", nil)
            if err == nil && pair != nil {
                contractAddr = string(pair.Value)
                break
            }
        }
        if i%2 == 0 { log.Println("SESSION: Aguardando contrato...") }
        time.Sleep(2 * time.Second)
    }

    if contractAddr != "" {
        var err error
        // MODO CONNECT: Passamos o endereço encontrado
        bcClient, _, err = blockchain.InitBlockchain(contractAddr)
        if err != nil {
            log.Printf("SESSION AVISO: Erro ao conectar no contrato %s: %v", contractAddr, err)
        } else {
            log.Printf("SESSION: Conectado com sucesso ao contrato compartilhado: %s", contractAddr)
        }
    } else {
        log.Println("SESSION AVISO: Timeout aguardando contrato. Auditoria desabilitada.")
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

	h.httpClient = &http.Client{ Timeout: 10 * time.Second }
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
	// O helper purchasePacksFromShop já envia o ID para o Shop registrar na blockchain
	initialCardKeys, err := h.purchasePacksFromShop(session.ID, initialPacksToOpen)
	if err != nil {
		log.Printf("CRITICAL: Failed to grant initial packs to player %s: %v", c.Conn().RemoteAddr(), err)
		welcomeMsg := "Welcome to the Jokenpo Game!\n\nCould not grant initial packs due to shop error."
		message.SendSuccessAndPrompt(c, state_LOBBY, "Connection successful!", welcomeMsg)
		return
	}

	for _, key := range initialCardKeys {
		if err := session.Player.AddCardToCollection(key, 1); err != nil {
			message.SendSuccessAndPrompt(c, state_LOBBY, "Connection successful, but error adding cards", err.Error())
			return
		}
	}

	deckBuildMessage := fmt.Sprintf("All %d initial cards added to collection/deck.", len(initialCardKeys))
	for i, key := range initialCardKeys {
		if _, err := session.Player.AddCardToDeck(key); err != nil {
			deckBuildMessage = fmt.Sprintf("Error building deck after %d cards: %v", i, err)
			break
		}
	}

	var sb strings.Builder
	sb.WriteString("Welcome to the Jokenpo Game!\n")
	sb.WriteString(fmt.Sprintf("As a bonus, you received %d card packs:\n", initialPacksToOpen))
	for i, key := range initialCardKeys {
		sb.WriteString(fmt.Sprintf("[%d] - %s \n", i, key))
	}
	sb.WriteString("\n" + deckBuildMessage)

	message.SendSuccessAndPrompt(c, state_LOBBY, "Connection successful! Welcome!", sb.String())
}

func (h *GameHandler) OnDisconnect(c *network.Client) {
	session, ok := h.sessionsByClient[c]
	if !ok { return }

	if session.State == state_IN_MATCH_QUEUE {
		h.leaveMatchQueue(session)
	} else if session.State == state_IN_TRADE_QUEUE {
		h.leaveTradeQueue(session)
	}

	delete(h.sessionsByClient, c)
	delete(h.sessionsByID, session.ID)
	log.Printf("Session %s removed.", session.ID)
}

func (h *GameHandler) OnMessage(c *network.Client, msg network.Message) {
	session, ok := h.sessionsByClient[c]
	if !ok { return }

	var router map[string]CommandHandlerFunc
	switch session.State {
	case state_LOBBY: router = h.lobbyRouter
	case state_IN_MATCH: router = h.matchRouter
	case state_IN_MATCH_QUEUE: router = h.matchQueueRouter
	case state_IN_TRADE_QUEUE: router = h.tradeQueueRouter
	default:
		message.SendErrorAndPrompt(c, "Invalid player state: %s", session.State)
		return
	}

	handler, found := router[msg.Type]
	if !found {
		message.SendErrorAndPrompt(c, "Unknown command: %s", msg.Type)
		return
	}
	handler(h, session, msg.Payload)
}
//END OF FILE jokenpo/internal/session/handler.go