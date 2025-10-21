//START OF FILE jokenpo/internal/session/handlers_lobby.go

package session

import (
	"encoding/json"
	"fmt"
	"jokenpo/internal/game/deck"
	"jokenpo/internal/session/message"
	"strings"
)

//Opção 1
func handleFindMatch(h *GameHandler, session *PlayerSession, payload json.RawMessage) {
	// 1. Validação de Estado
	if !checkLobbyState(session) {
		message.SendErrorAndPrompt(session.Client, "You are not in the lobby.")
		return
	}

	// 2. Chamada ao Serviço Externo
	err := h.enterMatchQueue(session)
	if err != nil {
		message.SendErrorAndPrompt(session.Client, "Failed to join match queue: %v", err)
		return
	}

	// 3. Atualização de Estado Local
	session.State = state_IN_MATCH_QUEUE
	
	// 4. Resposta ao Cliente
	message.SendSuccessAndPrompt(session.Client, session.State, "You have been added...", nil)
}

// handleTradeCard processa o comando do jogador para entrar na fila de troca às cegas.
func handleTradeCard(h *GameHandler, session *PlayerSession, payload json.RawMessage) {
	// 1. Validação de Contexto: Só pode trocar se estiver no lobby.
	if !checkLobbyState(session) {
		message.SendErrorAndPrompt(session.Client, "You must be in the lobby to trade a card.")
		return
	}

	// 2. Desserializa o Input: Precisamos saber qual carta o jogador quer oferecer.
	var req struct {
		CardKey string `json:"cardKey"`
	}
	if err := json.Unmarshal(payload, &req); err != nil || req.CardKey == "" {
		message.SendErrorAndPrompt(session.Client, "Invalid payload: 'cardKey' is required.")
		return
	}

	// 3. Validação de Negócio: O jogador realmente possui a carta que quer trocar?
	// (Esta é uma verificação importante para evitar trapaças).
	if err := session.Player.Inventory().HasCardInCollection(req.CardKey, 1); err != nil {
		message.SendErrorAndPrompt(session.Client, "Cannot trade card: %v", err)
		return
	}

	// 4. Lógica de Troca (Chamada ao Helper)
	// Removemos a carta da coleção do jogador ANTES de entrar na fila.
	// Se a entrada na fila falhar, precisaremos adicioná-la de volta (rollback).
	if err := session.Player.Inventory().Collection().RemoveCard(req.CardKey, 1); err != nil {
		message.SendErrorAndPrompt(session.Client, "An internal error occurred while preparing the trade: %v", err)
		return
	}

	err := h.enterTradeQueue(session, req.CardKey)
	if err != nil {
		// ROLLBACK: A chamada ao QueueService falhou. Devolvemos a carta ao jogador.
		session.Player.Inventory().Collection().AddCard(req.CardKey, 1)
		message.SendErrorAndPrompt(session.Client, "Failed to join trade queue: %v", err)
		return
	}

	// 5. Atualização de Estado e Resposta de Sucesso
	session.State = state_IN_TRADE_QUEUE // Novo estado
	
	message.SendSuccessAndPrompt(
		session.Client,
		session.State,
		fmt.Sprintf("You have entered the Wonder Trade queue, offering your '%s'.", req.CardKey),
		"Waiting for another player to trade with...",
	)
}

//Opção 3
// handlePurchasePackage processa o comando explícito do jogador para comprar pacotes.
// (Localizado em internal/session/handler.go)
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

	// 3. Lógica de Compra (Chamada ao Helper)
	cardKeys, err := h.purchasePacksFromShop(req.Quantity)
	if err != nil {
		// O helper retornou um erro (timeout, serviço offline, erro de negócio)
		// Aqui, tratamos o erro de forma contextual: informamos o jogador e pedimos um novo comando.
		message.SendErrorAndPrompt(session.Client, "Purchase failed: %v", err)
		return
	}

	// 4. Lógica de Negócio do Broker (Atualizar o Estado do Jogador)
	addedCards := []string{}
	for _, key := range cardKeys {
		if err := session.Player.Inventory().Collection().AddCard(key, 1); err != nil {
			// Se falhar a adição (ex: limite de coleção), é um erro do broker.
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
		"Package purchased successfully!",
		sb.String(),
	)
}


//Opção 4
func handleSeeCollection(h *GameHandler, session *PlayerSession, payload json.RawMessage) {

	if !checkLobbyState(session) {
		session.Client.Send() <- message.CreateErrorResponse("You are not in lobby")
		session.Client.Send() <- message.CreatePromptInputMessage()
		return
	}
	// 2. Chamada da Lógica de Negócio:
	collectionString, err := session.Player.SeeCollection()
	if err != nil {
		session.Client.Send() <- message.CreateErrorResponse(err.Error())
		session.Client.Send() <- message.CreatePromptInputMessage()
		return
	}

	// 3. Envio da Resposta de Sucesso:
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
	// 2. Chamada da Lógica de Negócio:
	// Acessamos o método 'SeeCollection' do Player, que já faz todo o trabalho de
	// buscar os dados do inventário e formatá-los em uma string legível.
	deckString, err := session.Player.SeeDeck()
	if err != nil {
		session.Client.Send() <- message.CreateErrorResponse(err.Error())
		session.Client.Send() <- message.CreatePromptInputMessage()
		return
	}

	// 3. Envio da Resposta de Sucesso:
	response := message.CreateSuccessResponse(session.State, "Your Deck:", deckString)
	session.Client.Send() <- response

	session.Client.Send() <- message.CreatePromptInputMessage()
}

// Adiciona um card da coleção pro deck, payload card key
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

//Remove uma carta do deck, payload int index
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

func (h *GameHandler) registerLobbyHandlers() {
	// --- Ações de Matchmaking ---
	// O comando para entrar na fila de espera para uma partida.
	h.lobbyRouter["FIND_MATCH"] = handleFindMatch
	h.lobbyRouter["TRADE_CARD"] = handleTradeCard

	// --- Ações da Loja ---
	// O comando para comprar pacotes de cartas.
	h.lobbyRouter["PURCHASE_PACKAGE"] = handlePurchasePackage

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


func checkLobbyState(session *PlayerSession) bool {
	return session.State == state_LOBBY
}

//END OF FILE jokenpo/internal/session/handlers_lobby.go
