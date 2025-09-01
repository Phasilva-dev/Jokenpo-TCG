package session

import (
	"encoding/json"
	"fmt"
	"jokenpo/internal/game/shop"
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
	shop       *shop.Shop

	roomFinished chan string

	// Teremos dois roteadores, um para cada estado do jogador.
	lobbyRouter map[string]CommandHandlerFunc
	matchRouter map[string]CommandHandlerFunc
}

// NewGameHandler agora também inicializa e registra os handlers dos roteadores.
func NewGameHandler() *GameHandler {
	h := &GameHandler{
		sessions:    make(map[*network.Client]*PlayerSession),
		matchmaker:  nil,
		rooms:       make(map[string]*GameRoom),
		shop:        shop.NewShop(),
		lobbyRouter: make(map[string]CommandHandlerFunc),
		matchRouter: make(map[string]CommandHandlerFunc),
	}
	h.matchmaker = NewMatchmaker(h)
	// Populamos os roteadores com seus respectivos comandos.
	h.registerLobbyHandlers()
	h.registerMatchHandlers()
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
	var purchasedPacksResults []string // Para guardar as strings formatadas originais
	var allObtainedCardKeys []string   // Para guardar apenas as chaves das cartas

	for i := 0; i < initialPacksToOpen; i++ {
		packResultStr, err := session.Player.PurchasePackage(h.shop)
		if err != nil {
			fmt.Printf("ERROR giving initial pack #%d to player %s: %v\n", i+1, c.Conn().RemoteAddr(), err)
			continue
		}
		
		purchasedPacksResults = append(purchasedPacksResults, packResultStr)
		
		cardKeys := parseCardKeysFromPackageString(packResultStr)
		allObtainedCardKeys = append(allObtainedCardKeys, cardKeys...)
	}

	// --- Lógica de Construção do Deck Inicial ---
	deckBuildMessage := "Your first 12 cards have been added to your deck."
	
	for _, key := range allObtainedCardKeys {
		_, err := session.Player.AddCardToDeck(key)
		if err != nil {
			deckBuildMessage = "Your initial cards were so powerful they exceeded the 80 power limit! Not all cards could be added to your starting deck."
			break
		}
	}

	// --- Formatação da Mensagem Final para o Cliente ---
	var sb strings.Builder
	sb.WriteString("Welcome to the Jokenpo Game!\n")
	sb.WriteString(fmt.Sprintf("As a bonus, you received %d card packs:\n\n", initialPacksToOpen))
	
	// Mostra os pacotes que o jogador abriu
	fullPacksString := strings.Join(purchasedPacksResults, "\n")
	sb.WriteString(fullPacksString)

	// Adiciona a mensagem sobre o status da construção do deck
	sb.WriteString("\n\n") // Duas quebras de linha para espaçamento
	sb.WriteString(deckBuildMessage)

	// Envia a resposta final
	welcomeMsg := message.CreateSuccessResponse(
		"Connection successful! Welcome!",
		sb.String(),
	)
	c.Send() <- welcomeMsg
}

func (h *GameHandler) OnDisconnect(c *network.Client) {
	// Futuramente: notificar sala de jogo, remover do matchmaking, etc.
	delete(h.sessions, c)
	fmt.Printf("Session removed. Total: %d\n", len(h.sessions))
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
	case StateLobby:
		router = h.lobbyRouter
	case StateInMatch:
		router = h.matchRouter
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