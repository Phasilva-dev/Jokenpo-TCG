package session

import (
	"encoding/json"
	"fmt"
	"jokenpo/internal/game/card"
	"jokenpo/internal/game/deck"
	"jokenpo/internal/session/message"
	"strings"
)

//Opção 1
func handleFindMatch(h *GameHandler, session *PlayerSession, payload json.RawMessage) {

	if checkLobbyState(session) {
		h.matchmaker.EnqueuePlayer(session)
		session.State = state_IN_QUEUE
	}

}

//Compra um pacote
//Opção 2
func handlePurchasePackage(h *GameHandler, session *PlayerSession, payload json.RawMessage) {

	result, err := session.Player.PurchasePackage(h.shop)
	if err != nil {
		session.Client.Send() <- message.CreateErrorResponse(err.Error())
		return
	}
	var sb strings.Builder
	sb.WriteString("The purchased cards are:\n")
	sb.WriteString(card.SliceOfCardsToString(result))
	session.Client.Send() <- message.CreateSuccessResponse("Package purchased successfully!", sb.String())
	
    printMenuClient(session)
}

// Compra multiplos pacotes (payload int amount)
//Opção 3
func handlePurchaseMultiPackage(h *GameHandler, session *PlayerSession, payload json.RawMessage) {
	var req struct {
		Amount *int `json:"amount"`
	}

	if err := json.Unmarshal(payload, &req); err != nil || req.Amount == nil {
		session.Client.Send() <- message.CreateErrorResponse("Invalid payload: 'amount' field is required and must be a number.")
		return
	}

	amount := *req.Amount

	if amount <= 0 || amount > 1000 {
		session.Client.Send() <- message.CreateErrorResponse("Invalid amount: must be between 1 and 1000.")
		return
	}

	allNewCards := []*card.Card{}

	for i := 0; i < amount; i++ {
		newCards, err := session.Player.PurchasePackage(h.shop)
		if err != nil {
			session.Client.Send() <- message.CreateErrorResponse(fmt.Sprintf("Failed on package #%d: %v", i+1, err))
			return
		}

		allNewCards = append(allNewCards, newCards...)
	}

	dataString := card.SliceOfCardsToString(allNewCards)

	successMessage := fmt.Sprintf("Successfully purchased %d packages!", amount)

	session.Client.Send() <- message.CreateSuccessResponse(successMessage, dataString)

	printMenuClient(session)

}

//Opção 4
func handleSeeCollection(h *GameHandler, session *PlayerSession, payload json.RawMessage) {

	// 2. Chamada da Lógica de Negócio:
	collectionString, err := session.Player.SeeCollection()
	if err != nil {
		session.Client.Send() <- message.CreateErrorResponse(err.Error())
		return
	}

	// 3. Envio da Resposta de Sucesso:
	response := message.CreateSuccessResponse("Your card collection:", collectionString)
	session.Client.Send() <- response

	printMenuClient(session)
}

//Opção 5
func handleSeeDeck(h *GameHandler, session *PlayerSession, payload json.RawMessage) {

	// 2. Chamada da Lógica de Negócio:
	// Acessamos o método 'SeeCollection' do Player, que já faz todo o trabalho de
	// buscar os dados do inventário e formatá-los em uma string legível.
	deckString, err := session.Player.SeeDeck()
	if err != nil {
		session.Client.Send() <- message.CreateErrorResponse(err.Error())
		return
	}

	// 3. Envio da Resposta de Sucesso:
	response := message.CreateSuccessResponse("Your Deck:", deckString)
	session.Client.Send() <- response

	printMenuClient(session)
}

// Adiciona um card da coleção pro deck, payload card key
//Opção 6
func handleAddCardToDeck(h *GameHandler, session *PlayerSession, payload json.RawMessage) {
	var req struct {
		Key *string `json:"key"`
	}

		if err := json.Unmarshal(payload, &req); err != nil || req.Key == nil {
		session.Client.Send() <- message.CreateErrorResponse("Invalid payload: 'key' field is required and must be a nom-empty string.")
		return
	}

	cardKey := *req.Key

	result, err := session.Player.AddCardToDeck(cardKey)

	if err != nil {
		session.Client.Send() <- message.CreateErrorResponse(err.Error())
		return
	}

	session.Client.Send() <- message.CreateSuccessResponse(result, nil)

	printMenuClient(session)
	
}

//Remove uma carta do deck, payload int index
//Opção 7
func handleRemoveCardFromDeck(h *GameHandler, session *PlayerSession, payload json.RawMessage) {
	var req struct {
		Index *int `json:"index"`
	}

	if err := json.Unmarshal(payload, &req); err != nil || req.Index == nil {
		session.Client.Send() <- message.CreateErrorResponse("Invalid payload: 'index' field is required and must be a number.")
		return
	}

	index := *req.Index
	deck, err := session.Player.Inventory().GameDeck().GetZone(deck.DECK)
	if err != nil {
		session.Client.Send() <- message.CreateErrorResponse(err.Error())
	}
	deckSize := deck.Size()

	if index < 0 || index > deckSize {
		session.Client.Send() <- message.CreateErrorResponse(fmt.Sprintf("Invalid index: must be between 0 and %d.", deckSize-1))
		return
	}

	result, err := session.Player.RemoveCardFromDeck(index)

	if err != nil {
		session.Client.Send() <- message.CreateErrorResponse(err.Error())
		return
	}
	
	session.Client.Send() <- message.CreateSuccessResponse(result, nil)

	printMenuClient(session)
	
}

// Opção 8
func handleReplaceCardToDeck(h *GameHandler, session *PlayerSession, payload json.RawMessage) {

	var req struct {
		IndexToRemove  *int    `json:"index"`
		KeyOfCardToAdd *string `json:"key"`
	}
	
	err := json.Unmarshal(payload, &req)
	if err != nil || req.IndexToRemove == nil || req.KeyOfCardToAdd == nil || *req.KeyOfCardToAdd == "" {
		const errorMessage = "Invalid payload: 'index' (a number) and 'key' (a non-empty string) are required."
		session.Client.Send() <- message.CreateErrorResponse(errorMessage)
		return
	}

	index := *req.IndexToRemove
	key := *req.KeyOfCardToAdd

	deck, err := session.Player.Inventory().GameDeck().GetZone(deck.DECK)
	if err != nil {
		session.Client.Send() <- message.CreateErrorResponse(err.Error())
	}
	deckSize := deck.Size()

	if index < 0 || index > deckSize {
		session.Client.Send() <- message.CreateErrorResponse(fmt.Sprintf("Invalid amount: must be between 0 and %d.", deckSize-1))
		return
	}

	result, err := session.Player.ReplaceCardInDeck(index,key)

	if err != nil {
		session.Client.Send() <- message.CreateErrorResponse(err.Error())
		return
	}
	
	session.Client.Send() <- message.CreateSuccessResponse(result, nil)

	printMenuClient(session)

}

func handleLeaveQueue(h *GameHandler, session *PlayerSession, payload json.RawMessage) {


	if checkQueueState(session) {
		//
		msg := message.CreateSuccessResponse("Your request to leave the queue was received.", nil)
		session.Client.Send() <- msg

		h.matchmaker.LeaveQueue(session)
		session.State = state_LOBBY
	}

}



func (h *GameHandler) registerLobbyHandlers() {
	// --- Ações de Matchmaking ---
	// O comando para entrar na fila de espera para uma partida.
	h.lobbyRouter["FIND_MATCH"] = handleFindMatch
	h.lobbyRouter["LEAVE_QUEUE"] = handleLeaveQueue

	// --- Ações da Loja ---
	// O comando para comprar um único pacote de cartas.
	h.lobbyRouter["PURCHASE_PACKAGE"] = handlePurchasePackage
	// O comando para comprar múltiplos pacotes de uma vez.
	h.lobbyRouter["PURCHASE_MULTI_PACKAGE"] = handlePurchaseMultiPackage

	// --- Ações de Visualização (Read-Only) ---
	// O comando para ver a coleção completa de cartas do jogador.
	h.lobbyRouter["VIEW_COLLECTION"] = handleSeeCollection
	// O comando para ver o baralho de jogo atual.
	h.lobbyRouter["VIEW_DECK"] = handleSeeDeck

	// --- Ações de Gerenciamento do Baralho ---
	// O comando para adicionar uma carta da coleção ao baralho.
	h.lobbyRouter["ADD_CARD_TO_DECK"] = handleAddCardToDeck
	// O comando para remover uma carta do baralho pelo seu índice.
	h.lobbyRouter["REMOVE_CARD_FROM_DECK"] = handleRemoveCardFromDeck
	// O comando para substituir uma carta no baralho por outra da coleção.
	h.lobbyRouter["REPLACE_CARD_TO_DECK"] = handleReplaceCardToDeck
}

func printMainMenu() string {
	var sb strings.Builder
	sb.WriteString("\n--- Jokenpo Card Game (Lobby) ---\n")
	sb.WriteString("1. Buscar Partida\n")
	sb.WriteString("2. Comprar Pacote\n")
	sb.WriteString("3. Comprar Múltiplos Pacotes\n")
	sb.WriteString("4. Ver Coleção\n")
	sb.WriteString("5. Ver Deck\n")
	sb.WriteString("6. Adicionar Carta ao Deck\n")
	sb.WriteString("7. Remover Carta do Deck\n")
	sb.WriteString("8. Substituir Carta no Deck\n")
	sb.WriteString("---------------------------------\n")
	sb.WriteString("Escolha uma opção: ")

	return sb.String()
}

func printMenuClient(session *PlayerSession) {
	session.Client.Send() <- message.CreateSuccessResponse(printMainMenu(),nil)
}

func checkLobbyState(session *PlayerSession) bool {
	return session.State == state_LOBBY
}

func checkQueueState(session *PlayerSession) bool {
	return session.State == state_IN_QUEUE
}
