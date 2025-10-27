//START OF FILE jokenpo/internal/session/handler.go
package session

import (
	"encoding/json"
	"fmt"
	"jokenpo/internal/network"
	"jokenpo/internal/services/cluster"
	"jokenpo/internal/session/message" // Importe seu novo pacote de mensagens
	"log"
	"net/http"
	"strings"
	"time"
)

// CommandHandlerFunc define a assinatura para todas as nossas funções que lidam com comandos.
// Elas recebem o contexto da sessão e o payload bruto da mensagem.
type CommandHandlerFunc func(h *GameHandler, session *PlayerSession, payload json.RawMessage)

type GameHandler struct {
	sessionsByClient map[*network.Client]*PlayerSession // Para lookups a partir da conexão
	sessionsByID     map[string]*PlayerSession        // Para lookups a partir do UUID (callbacks)

	httpClient *http.Client
	serviceCache *cluster.ServiceCacheActor
	advertisedHostname string

	// Teremos dois roteadores, um para cada estado do jogador.
	lobbyRouter map[string]CommandHandlerFunc
	matchRouter map[string]CommandHandlerFunc
	matchQueueRouter map[string]CommandHandlerFunc
	tradeQueueRouter map[string]CommandHandlerFunc
}

// NewGameHandler agora também inicializa e registra os handlers dos roteadores.
func NewGameHandler(consulAddr, advertisedHostname string) (*GameHandler, error) {
	h := &GameHandler{
		sessionsByClient: make(map[*network.Client]*PlayerSession),
		sessionsByID:     make(map[string]*PlayerSession),
		advertisedHostname: advertisedHostname,
		lobbyRouter:      make(map[string]CommandHandlerFunc),
		matchRouter:      make(map[string]CommandHandlerFunc),
		matchQueueRouter: make(map[string]CommandHandlerFunc),
		tradeQueueRouter: make(map[string]CommandHandlerFunc),
	}

	h.httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}
	h.serviceCache = cluster.NewServiceCacheActor(30 * time.Second, consulAddr)

	h.registerLobbyHandlers()
	h.registerQueueHandlers()
	h.registerMatchHandlers()

	return h, nil
}

// --- Implementação da Interface network.EventHandler ---

func (h *GameHandler) OnConnect(c *network.Client) {
	// 1. Cria a sessão do jogador
	session := NewPlayerSession(c)
	
	// Adiciona a sessão a ambos os mapas
	h.sessionsByClient[c] = session
	h.sessionsByID[session.ID] = session

	log.Printf("Session created for %s. Total sessions: %d", c.Conn().RemoteAddr(), len(h.sessionsByClient))

	// --- 2. Lógica de Compra Inicial ---
	const initialPacksToOpen = 4
	initialCardKeys, err := h.purchasePacksFromShop(initialPacksToOpen)

	if err != nil {
		log.Printf("CRITICAL: Failed to grant initial packs to player %s: %v", c.Conn().RemoteAddr(), err)
		welcomeMsg := "Welcome to the Jokenpo Game!\n\n" +
			"Unfortunately, we could not grant you your initial card packs at this time as our shop is unavailable. " +
			"Please try the 'purchase' command later."
		message.SendSuccessAndPrompt(c, state_LOBBY, "Connection successful!", welcomeMsg)
		return
	}

	// --- 3. Adiciona as Cartas à Coleção e ao Deck em Fases Separadas ---

	// FASE 3.1: Adicionar todas as cartas à coleção.
	for _, key := range initialCardKeys {
		if err := session.Player.AddCardToCollection(key, 1); err != nil {
			// Este é um erro grave e inesperado. Se falhar aqui, não podemos continuar.
			log.Printf("CRITICAL ERROR: Failed to add purchased card '%s' to collection for player %s: %v", key, c.Conn().RemoteAddr(), err)
			
			// Formata a mensagem de boas-vindas com o erro e sai.
			var sb strings.Builder
			sb.WriteString("Welcome to the Jokenpo Game!\n")
			sb.WriteString(fmt.Sprintf("You received %d card packs, but a critical error occurred while adding them to your collection:\n\n", initialPacksToOpen))
			sb.WriteString(fmt.Sprintf("Error: %v\n\n", err))
			sb.WriteString("Please contact support.")

			message.SendSuccessAndPrompt(c, state_LOBBY, "Connection successful, but an error occurred!", sb.String())
			return
		}
	}
	
	// Prepara a mensagem de construção do deck. Começa com uma mensagem de sucesso.
	deckBuildMessage := fmt.Sprintf("All %d initial cards have been added to your collection and starting deck.", len(initialCardKeys))

	// FASE 3.2: Adicionar todas as cartas ao deck inicial.
	for i, key := range initialCardKeys {
		if _, err := session.Player.AddCardToDeck(key); err != nil {
			// Este é um erro esperado (ex: limite de poder do deck).
			// O jogador já tem as cartas na coleção, então apenas o informamos.
			deckBuildMessage = fmt.Sprintf(
				"All cards were added to your collection, but an error occurred while building your starting deck after adding %d cards.\nReason: %v",
				i, err,
			)
			log.Printf("INFO: Could not add card '%s' to initial deck for player %s: %v", key, c.Conn().RemoteAddr(), err)
			break // Interrompe a construção do deck, mas a operação geral é um "sucesso parcial".
		}
	}

	// --- 4. Formata e Envia a Mensagem Final de Boas-Vindas ---
	var sb strings.Builder
	sb.WriteString("Welcome to the Jokenpo Game!\n")
	sb.WriteString(fmt.Sprintf("As a bonus, you received %d card packs, revealing the following cards:\n\n", initialPacksToOpen))
	
	// Lista as chaves das cartas que o jogador abriu.
	for i, key := range initialCardKeys {
		msg := fmt.Sprintf("[%d] - %s \n", i, key)
		sb.WriteString(msg)
	}

	sb.WriteString("\n") // Espaçamento
	
	// Adiciona a mensagem final sobre o status da coleção/deck.
	sb.WriteString(deckBuildMessage)

	// Envia a resposta completa para o cliente.
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

	// Notifica o QueueService se o jogador estava em uma fila.
	if session.State == state_IN_MATCH_QUEUE {
		log.Printf("Player %s disconnected while in match queue. Notifying Queue Service.", session.ID)
		h.leaveMatchQueue(session) // Chama o helper de API para sair
	} else if session.State == state_IN_TRADE_QUEUE {
		log.Printf("Player %s disconnected while in trade queue. Notifying Queue Service.", session.ID)
		h.leaveTradeQueue(session) // Chama o helper de API para sair
	}

	// Remove a sessão de ambos os mapas
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