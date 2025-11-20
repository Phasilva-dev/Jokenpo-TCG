//START OF FILE jokenpo/internal/session/handlers_lobby.go
package session

import (
	"encoding/json"
	"fmt"
	"jokenpo/internal/game/deck"
	"jokenpo/internal/session/message"
	"strings"
)

// ... (handleFindMatch e handleTradeCard permanecem inalterados) ...

//Opção 1
func handleFindMatch(h *GameHandler, session *PlayerSession, payload json.RawMessage) {
	if !checkLobbyState(session) {
		message.SendErrorAndPrompt(session.Client, "You are not in the lobby.")
		return
	}

	deckJSON, err := session.Player.Inventory().GameDeck().ToJSON()
	if err != nil {
		message.SendErrorAndPrompt(session.Client, "Failed to prepare your deck for matchmaking: %v", err)
		return
	}
	var deckKeys []string
	if err := json.Unmarshal(deckJSON, &deckKeys); err != nil {
		message.SendErrorAndPrompt(session.Client, "Failed to process your deck for matchmaking: %v", err)
		return
	}

	err = h.enterMatchQueue(session, deckKeys)
	if err != nil {
		message.SendErrorAndPrompt(session.Client, "Failed to join match queue: %v", err)
		return
	}

	session.State = state_IN_MATCH_QUEUE
	message.SendSuccessAndPrompt(session.Client, session.State, "You have been added to the matchmaking queue. Searching for an opponent...", nil)
}

//Opção 2
func handleTradeCard(h *GameHandler, session *PlayerSession, payload json.RawMessage) {
	if !checkLobbyState(session) {
		message.SendErrorAndPrompt(session.Client, "You must be in the lobby to trade a card.")
		return
	}

	var req struct {
		CardKey string `json:"cardKey"`
	}
	if err := json.Unmarshal(payload, &req); err != nil || req.CardKey == "" {
		message.SendErrorAndPrompt(session.Client, "Invalid payload: 'cardKey' is required.")
		return
	}

	if err := session.Player.Inventory().HasCardInCollection(req.CardKey, 1); err != nil {
		message.SendErrorAndPrompt(session.Client, "Cannot trade card: %v", err)
		return
	}

	if err := session.Player.Inventory().Collection().RemoveCard(req.CardKey, 1); err != nil {
		message.SendErrorAndPrompt(session.Client, "An internal error occurred while preparing the trade: %v", err)
		return
	}

	err := h.enterTradeQueue(session, req.CardKey)
	if err != nil {
		session.Player.Inventory().Collection().AddCard(req.CardKey, 1)
		message.SendErrorAndPrompt(session.Client, "Failed to join trade queue: %v", err)
		return
	}

	session.State = state_IN_TRADE_QUEUE
	message.SendSuccessAndPrompt(
		session.Client,
		session.State,
		fmt.Sprintf("You have entered the Wonder Trade queue, offering your '%s'.", req.CardKey),
		"Waiting for another player to trade with...",
	)
}

//Opção 3
// handlePurchasePackage processa o comando explícito do jogador para comprar pacotes.
func handlePurchasePackage(h *GameHandler, session *PlayerSession, payload json.RawMessage) {
	// 1. Validação de Contexto
	if !checkLobbyState(session) {
		message.SendErrorAndPrompt(session.Client, "You are not in lobby")
		return
	}

	// 2. Desserializa o Input
	var req struct {
		Quantity uint64 `json:"quantity"`
	}
	if err := json.Unmarshal(payload, &req); err != nil || req.Quantity == 0 {
		message.SendErrorAndPrompt(session.Client, "Invalid payload: 'quantity' is required and must be a number")
		return
	}

	// 3. Lógica de Compra (Chamada ao Shop Service)
	// O Shop agora é responsável por gerar UUIDs e registrar na Blockchain!
	cardKeys, err := h.purchasePacksFromShop(session.ID, req.Quantity)
	if err != nil {
		message.SendErrorAndPrompt(session.Client, "Purchase failed: %v", err)
		return
	}

	// 4. Lógica de Negócio do Broker (Atualizar o Estado do Jogador)
	addedCards := []string{}
	for _, key := range cardKeys {
		if err := session.Player.Inventory().Collection().AddCard(key, 1); err != nil {
			message.SendErrorAndPrompt(session.Client, "Failed to add card '%s' to your collection: %v", key, err)
			return
		}
		addedCards = append(addedCards, key)
	}

	// 5. Formatação e Resposta de Sucesso
	var sb strings.Builder
	sb.WriteString("The purchased cards are:\n")
	for _, key := range addedCards {
		sb.WriteString("- " + key + "\n")
	}

	message.SendSuccessAndPrompt(
		session.Client,
		session.State,
		"Package purchased successfully! (Registered on Blockchain)",
		sb.String(),
	)
}

// ... (Resto dos handlers sem alterações) ...

//Opção 4
func handleSeeCollection(h *GameHandler, session *PlayerSession, payload json.RawMessage) {

	if !checkLobbyState(session) {
		session.Client.Send() <- message.CreateErrorResponse("You are not in lobby")
		session.Client.Send() <- message.CreatePromptInputMessage()
		return
	}
	collectionString, err := session.Player.SeeCollection()
	if err != nil {
		session.Client.Send() <- message.CreateErrorResponse(err.Error())
		session.Client.Send() <- message.CreatePromptInputMessage()
		return
	}
	response := message.CreateSuccessResponse(session.State, "Your card collection:", collectionString)
	session.Client.Send() <- response
	session.Client.Send() <- message.CreatePromptInputMessage()
}

//Opção 5
func handleSeeDeck(h *GameHandler, session *PlayerSession, payload json.RawMessage) {

	if !checkLobbyState(session) {
		session.Client.Send() <- message.CreateErrorResponse("You are not in lobby")
		session.Client.Send() <- message.CreatePromptInputMessage()
		return
	}
	deckString, err := session.Player.SeeDeck()
	if err != nil {
		session.Client.Send() <- message.CreateErrorResponse(err.Error())
		session.Client.Send() <- message.CreatePromptInputMessage()
		return
	}
	response := message.CreateSuccessResponse(session.State, "Your Deck:", deckString)
	session.Client.Send() <- response
	session.Client.Send() <- message.CreatePromptInputMessage()
}

//Opção 6
func handleAddCardToDeck(h *GameHandler, session *PlayerSession, payload json.RawMessage) {

	if !checkLobbyState(session) {
		session.Client.Send() <- message.CreateErrorResponse("You are not in lobby")
		session.Client.Send() <- message.CreatePromptInputMessage()
		return
	}

	var req struct {
		Key *string `json:"key"`
	}

		if err := json.Unmarshal(payload, &req); err != nil || req.Key == nil {
		session.Client.Send() <- message.CreateErrorResponse("Invalid payload: 'key' field is required and must be a nom-empty string.")
		session.Client.Send() <- message.CreatePromptInputMessage()
		return
	}

	cardKey := *req.Key
	result, err := session.Player.AddCardToDeck(cardKey)

	if err != nil {
		session.Client.Send() <- message.CreateErrorResponse(err.Error())
		session.Client.Send() <- message.CreatePromptInputMessage()
		return
	}
	session.Client.Send() <- message.CreateSuccessResponse(session.State, result, nil)
	session.Client.Send() <- message.CreatePromptInputMessage()
}

//Opção 7
func handleRemoveCardFromDeck(h *GameHandler, session *PlayerSession, payload json.RawMessage) {

	if !checkLobbyState(session) {
		session.Client.Send() <- message.CreateErrorResponse("You are not in lobby")
		session.Client.Send() <- message.CreatePromptInputMessage()
		return
	}

	var req struct {
		Index *int `json:"index"`
	}

	if err := json.Unmarshal(payload, &req); err != nil || req.Index == nil {
		session.Client.Send() <- message.CreateErrorResponse("Invalid payload: 'index' field is required and must be a number.")
		session.Client.Send() <- message.CreatePromptInputMessage()
		return
	}

	index := *req.Index
	deck, err := session.Player.Inventory().GameDeck().GetZone(deck.DECK)
	if err != nil {
		session.Client.Send() <- message.CreateErrorResponse(err.Error())
		session.Client.Send() <- message.CreatePromptInputMessage()
	}
	deckSize := deck.Size()
	if deckSize == 8 {
		session.Client.Send() <- message.CreateErrorResponse("You cannot have less than 8 cards in deck")
		session.Client.Send() <- message.CreatePromptInputMessage()
		return
	}

	if index < 0 || index > deckSize {
		session.Client.Send() <- message.CreateErrorResponse(fmt.Sprintf("Invalid index: must be between 0 and %d.", deckSize-1))
		session.Client.Send() <- message.CreatePromptInputMessage()
		return
	}

	result, err := session.Player.RemoveCardFromDeck(index)

	if err != nil {
		session.Client.Send() <- message.CreateErrorResponse(err.Error())
		session.Client.Send() <- message.CreatePromptInputMessage()
		return
	}
	session.Client.Send() <- message.CreateSuccessResponse(session.State, result, nil)
	session.Client.Send() <- message.CreatePromptInputMessage()
}

// Opção 8
func handleReplaceCardToDeck(h *GameHandler, session *PlayerSession, payload json.RawMessage) {

	if !checkLobbyState(session) {
		session.Client.Send() <- message.CreateErrorResponse("You are not in lobby")
		session.Client.Send() <- message.CreatePromptInputMessage()
		return
	}

	var req struct {
		IndexToRemove  *int    `json:"index"`
		KeyOfCardToAdd *string `json:"key"`
	}
	
	err := json.Unmarshal(payload, &req)
	if err != nil || req.IndexToRemove == nil || req.KeyOfCardToAdd == nil || *req.KeyOfCardToAdd == "" {
		const errorMessage = "Invalid payload: 'index' (a number) and 'key' (a non-empty string) are required."
		session.Client.Send() <- message.CreateErrorResponse(errorMessage)
		session.Client.Send() <- message.CreatePromptInputMessage()
		return
	}

	index := *req.IndexToRemove
	key := *req.KeyOfCardToAdd

	deck, err := session.Player.Inventory().GameDeck().GetZone(deck.DECK)
	if err != nil {
		session.Client.Send() <- message.CreateErrorResponse(err.Error())
		session.Client.Send() <- message.CreatePromptInputMessage()
	}
	deckSize := deck.Size()

	if index < 0 || index > deckSize {
		session.Client.Send() <- message.CreateErrorResponse(fmt.Sprintf("Invalid amount: must be between 0 and %d.", deckSize-1))
		session.Client.Send() <- message.CreatePromptInputMessage()
		return
	}

	result, err := session.Player.ReplaceCardInDeck(index,key)

	if err != nil {
		session.Client.Send() <- message.CreateErrorResponse(err.Error())
		session.Client.Send() <- message.CreatePromptInputMessage()
		return
	}
	session.Client.Send() <- message.CreateSuccessResponse(session.State, result, nil)
	session.Client.Send() <- message.CreatePromptInputMessage()

}

//Opção 10
func handleViewAuditLogs(h *GameHandler, session *PlayerSession, payload json.RawMessage) {
	if !checkLobbyState(session) {
		message.SendErrorAndPrompt(session.Client, "You are not in lobby")
		return
	}

    if h.blockchain == nil {
        message.SendErrorAndPrompt(session.Client, "Blockchain service is currently unavailable.")
        return
    }

    report, err := h.blockchain.GetAuditReport()
    if err != nil {
        message.SendErrorAndPrompt(session.Client, "Failed to fetch audit logs: %v", err)
        return
    }

	message.SendSuccessAndPrompt(
		session.Client,
		session.State,
		"Audit Report fetched from Blockchain:",
		report, 
	)
}

func (h *GameHandler) registerLobbyHandlers() {
	h.lobbyRouter["FIND_MATCH"] = handleFindMatch
	h.lobbyRouter["TRADE_CARD"] = handleTradeCard
	h.lobbyRouter["PURCHASE_PACKAGE"] = handlePurchasePackage
	h.lobbyRouter["VIEW_COLLECTION"] = handleSeeCollection
	h.lobbyRouter["VIEW_DECK"] = handleSeeDeck
	h.lobbyRouter["ADD_CARD_TO_DECK"] = handleAddCardToDeck
	h.lobbyRouter["REMOVE_CARD_FROM_DECK"] = handleRemoveCardFromDeck
	h.lobbyRouter["REPLACE_CARD_TO_DECK"] = handleReplaceCardToDeck
	h.lobbyRouter["VIEW_AUDIT"] = handleViewAuditLogs
}

func checkLobbyState(session *PlayerSession) bool {
	return session.State == state_LOBBY
}
//END OF FILE jokenpo/internal/session/handlers_lobby.go