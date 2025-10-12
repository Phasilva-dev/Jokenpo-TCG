package session

import (
	"encoding/json"
	"fmt"
	"jokenpo/internal/network"
	"jokenpo/internal/services/cluster"
	"jokenpo/internal/session/message" // Importe seu novo pacote de mensagens
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
func NewGameHandler() *GameHandler {
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

	h.serviceCache = cluster.NewServiceCacheActor(30 * time.Second)

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

	// --- 2. Lógica de Compra Inicial (Chamada ao Helper) ---
	const initialPacksToOpen = 4
	initialCardKeys, err := h.purchasePacksFromShop(initialPacksToOpen)

	// Tratamento de Erro CONTEXTUAL do OnConnect: Loga erro crítico e envia mensagem alternativa.
	if err != nil {
		fmt.Printf("CRITICAL: Failed to grant initial packs to player %s: %v\n", c.Conn().RemoteAddr(), err)

		// Envia mensagem de boas-vindas informando sobre o problema, mas permitindo a conexão.
		welcomeMsg := "Welcome to the Jokenpo Game!\n\n" +
			"Unfortunately, we could not grant you your initial card packs at this time as our shop is unavailable. " +
			"Please try the 'purchase' command later."

		message.SendSuccessAndPrompt(c, state_LOBBY, "Connection successful!", welcomeMsg)
		return // Interrompe o fluxo de onboarding se a compra falhar.
	}

	// --- 3. Adiciona as Cartas Recebidas ao Deck Inicial ---
	deckBuildMessage := fmt.Sprintf("Your first %d cards have been added to your deck.", len(initialCardKeys))
	
	for i, key := range initialCardKeys {
		// A lógica interna do Player (AddCardToDeck) permanece a mesma.
		_, err := session.Player.AddCardToDeck(key)
		if err != nil {
			deckBuildMessage = fmt.Sprintf("Your initial cards were so powerful they exceeded the 80 power limit!\n"+
				"Not all cards could be added to your starting deck.\n"+
				"You have added only %d cards.", i)
			break // Para de adicionar cartas se o limite for atingido.
		}
	}

	// --- 4. Formata e Envia a Mensagem Final de Boas-Vindas ---
	var sb strings.Builder
	sb.WriteString("Welcome to the Jokenpo Game!\n")
	sb.WriteString(fmt.Sprintf("As a bonus, you received %d card packs, revealing the following cards:\n\n", initialPacksToOpen))
	
	// Lista as chaves das cartas que o jogador abriu
	for _, key := range initialCardKeys {
		sb.WriteString("- " + key + "\n")
	}

	sb.WriteString("\n") // Espaçamento
	sb.WriteString(deckBuildMessage)

	// Usa o helper para enviar a mensagem de sucesso e o prompt de uma só vez.
	message.SendSuccessAndPrompt(
		c, // O network.Client satisfaz a interface MessageSender
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