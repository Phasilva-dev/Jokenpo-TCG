package session

import (
	"encoding/json"
	"fmt"
	"jokenpo/internal/game/card"
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
	sessions   map[*network.Client]*PlayerSession

	httpClient *http.Client
	serviceCache *cluster.ServiceCacheActor

	matchmaker *Matchmaker
	rooms      map[string]*GameRoom

	roomFinished chan string

	// Teremos dois roteadores, um para cada estado do jogador.
	lobbyRouter map[string]CommandHandlerFunc
	matchRouter map[string]CommandHandlerFunc
	queueRouter map[string]CommandHandlerFunc
}

// NewGameHandler agora também inicializa e registra os handlers dos roteadores.
func NewGameHandler(consulAddr string) (*GameHandler, error) {
	h := &GameHandler{
		sessions:    make(map[*network.Client]*PlayerSession),
		matchmaker:  nil, // será inicializado abaixo
		rooms:       make(map[string]*GameRoom),
		lobbyRouter: make(map[string]CommandHandlerFunc),
		matchRouter: make(map[string]CommandHandlerFunc),
		queueRouter: make(map[string]CommandHandlerFunc),
	}
	h.matchmaker = NewMatchmaker(h)

	// Inicializa o cliente HTTP compartilhado para toda a comunicação com microsserviços.
	// Usar um único cliente com um timeout é crucial para performance e estabilidade.
	h.httpClient = &http.Client{
		Timeout: 10 * time.Second, // Timeout de 10 segundos para evitar que uma chamada prenda a goroutine do jogador
	}

	h.serviceCache = cluster.NewServiceCacheActor(30 * time.Second, consulAddr)

	// Populamos os roteadores com seus respectivos comandos.
	h.registerLobbyHandlers()
	h.registerMatchHandlers()
	h.registerQueueHandlers()
	err := card.InitGlobalCatalog()
	if err != nil {
		return nil, err
	}
	return h, nil
}

// --- Implementação da Interface network.EventHandler ---

func (h *GameHandler) OnConnect(c *network.Client) {
	// 1. Cria a sessão do jogador
	session := NewPlayerSession(c)
	h.sessions[c] = session
	log.Printf("Session created for %s. Total sessions: %d", c.Conn().RemoteAddr(), len(h.sessions))

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
	// 1. Encontra a sessão do cliente que desconectou.
	session, ok := h.sessions[c]
	if !ok {
		// Se não havia sessão, não há nada para limpar.
		return
	}

	// 2. LÓGICA DE LIMPEZA CENTRAL: Verifica o estado do jogador.
	// Esta é a correção para o bug.
	switch session.State {
	case state_IN_QUEUE:
		// Se o jogador estava na fila, avisa o Matchmaker para removê-lo.
		// Isso previne que o Matchmaker tente enviar mensagens para um canal fechado.
		fmt.Printf("Player %s disconnected while in queue. Removing from matchmaking.\n", c.Conn().RemoteAddr())
		h.matchmaker.LeaveQueue(session)

	case state_IN_MATCH:
		// Se o jogador estava em uma partida, avisa a GameRoom.
		// A GameRoom então lidará com a lógica de fim de jogo por desconexão.
		if session.CurrentRoom != nil {
			fmt.Printf("Player %s disconnected from room %s.\n", c.Conn().RemoteAddr(), session.CurrentRoom.ID)
			session.CurrentRoom.unregister <- session
		}
	}

	// 3. Após notificar os outros sistemas, remove a sessão do mapa principal.
	delete(h.sessions, c)
	fmt.Printf("Session for %s removed. Total sessions: %d\n", c.Conn().RemoteAddr(), len(h.sessions))
}

// OnMessage agora é um despachante limpo e simples.
func (h *GameHandler) OnMessage(c *network.Client, msg network.Message) {
	session, ok := h.sessions[c]
	if !ok {
		return // Ignora mensagens de clientes sem sessão.
	}

	var router map[string]CommandHandlerFunc
	// 1. Seleciona o roteador apropriado baseado no estado do jogador.
	switch session.State {
	case state_LOBBY:
		router = h.lobbyRouter
	case state_IN_MATCH:
		router = h.matchRouter
	case state_IN_QUEUE:
		router = h.queueRouter
	default:
		c.Send() <- message.CreateErrorResponse(fmt.Sprintf("Invalid state of player: %s", session.State))
		return
	}

	// 2. Procura pelo handler do comando no roteador selecionado.
	handler, found := router[msg.Type]
	if !found {
		c.Send() <- message.CreateErrorResponse(fmt.Sprintf("Unknown or invalid command for actual state of player: %s", msg.Type))
		return
	}

	// 3. Executa o handler encontrado.
	handler(h, session, msg.Payload)
}