package session

import (
	"encoding/json"
	"fmt"
	"jokenpo/internal/game/card"
	"jokenpo/internal/game/deck"
	"jokenpo/internal/session/message"
	"strings"
)

//Compra um pacote
func handlePurchasePackage(h *GameHandler, session *PlayerSession, payload json.RawMessage) {

	result, err := session.Player.PurchasePackage(h.shop)
	if err != nil {
		session.Client.Send() <- message.CreateErrorResponse(err.Error())
		return
	}
	var sb strings.Builder
	sb.WriteString("The purchased cards are:\n")
	sb.WriteString(card.SliceOfCardsToString(result))
	session.Client.Send() <- message.CreateSuccessResponse("Package purchased successfully!", result)
}

// Compra multiplos pacotes (payload int amount)
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

}

// Adiciona um card da coleção pro deck, payload card key
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
	
}

//Remove uma carta do deck, payload int index
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
	
}

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

}

func handleFindMatch(h *GameHandler, session *PlayerSession, payload json.RawMessage) {

	h.matchmaker.EnqueuePlayer(session)

}

func handleLeaveQueue(h *GameHandler, session *PlayerSession, payload json.RawMessage) {
	h.matchmaker.LeaveQueue(session)
}

// Read-only

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
}

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
