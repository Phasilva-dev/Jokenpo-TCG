package session

import (
	"encoding/json"
	"fmt"
	"jokenpo/internal/game/card"
	"jokenpo/internal/network"
	"jokenpo/internal/session/message" // Importe seu novo pacote de mensagens
	"strings"
)

// CommandHandlerFunc define a assinatura para todas as nossas funções que lidam com comandos.
// Elas recebem o contexto da sessão e o payload bruto da mensagem.
type CommandHandlerFunc func(h *GameHandler, session *PlayerSession, payload json.RawMessage)

type GameHandler struct {
	sessions   map[*network.Client]*PlayerSession
	matchmaker *Matchmaker
	rooms      map[string]*GameRoom

	roomFinished chan string

	// Teremos dois roteadores, um para cada estado do jogador.
	lobbyRouter map[string]CommandHandlerFunc
	matchRouter map[string]CommandHandlerFunc
	queueRouter map[string]CommandHandlerFunc
}

// NewGameHandler agora também inicializa e registra os handlers dos roteadores.
func NewGameHandler() *GameHandler {
	h := &GameHandler{
		sessions:    make(map[*network.Client]*PlayerSession),
		matchmaker:  nil,
		rooms:       make(map[string]*GameRoom),
		lobbyRouter: make(map[string]CommandHandlerFunc),
		matchRouter: make(map[string]CommandHandlerFunc),
	}
	h.matchmaker = NewMatchmaker(h)
	// Populamos os roteadores com seus respectivos comandos.
	h.registerLobbyHandlers()
	h.registerMatchHandlers()
	h.registerQueueHandlers()
	return h
}

// --- Implementação da Interface network.EventHandler ---

// OnConnect é chamado pela goroutine do network.Hub. É seguro modificar o estado aqui.
func (h *GameHandler) OnConnect(c *network.Client) {
	// 1. Cria a sessão do jogador
	session := NewPlayerSession(c)
	h.sessions[c] = session
	fmt.Printf("Session created for %s. Total sessions: %d\n", c.Conn().RemoteAddr(), len(h.sessions))

	// --- Abertura de Pacotes ---
	const initialPacksToOpen = 4
	var purchasedPacksResults []*card.Card // Para guardar as strings formatadas originais

	/*
	for i := 0; i < initialPacksToOpen; i++ {
		packResult, err := session.Player.PurchasePackage(h.shop)
		if err != nil {
			fmt.Printf("ERROR giving initial pack #%d to player %s: %v\n", i+1, c.Conn().RemoteAddr(), err)
			continue
		}
		
		purchasedPacksResults = append(purchasedPacksResults, packResult...)
	
	}*/

	// --- Lógica de Construção do Deck Inicial ---
	deckBuildMessage := "Your first 12 cards have been added to your deck."
	
	for i, card := range purchasedPacksResults {
		_, err := session.Player.AddCardToDeck(card.Key())
		if err != nil {
			
			deckBuildMessage = 
			fmt.Sprintf("Your initial cards were so powerful they exceeded the 80 power limit!\n Not all cards could be added to your starting deck.\n You have added only %d", i)
			break
		}
	}

	// --- Formatação da Mensagem Final para o Cliente ---
	var sb strings.Builder
	sb.WriteString("Welcome to the Jokenpo Game!\n")
	sb.WriteString(fmt.Sprintf("As a bonus, you received %d card packs:\n\n", initialPacksToOpen))
	
	// Mostra os pacotes que o jogador abriu
	results := card.SliceOfCardsToString(purchasedPacksResults)
	sb.WriteString(results)

	// Adiciona a mensagem sobre o status da construção do deck
	sb.WriteString("\n\n") // Duas quebras de linha para espaçamento
	sb.WriteString(deckBuildMessage)

	// Envia a resposta final
	welcomeMsg := message.CreateSuccessResponse(state_LOBBY,
		"Connection successful! Welcome!",
		sb.String(),
	)
	c.Send() <- welcomeMsg

	c.Send() <- message.CreatePromptInputMessage()
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