// START OF FILE jokenpo/internal/services/gameroom/room_logic.go
package gameroom

import (
	"fmt"
	"jokenpo/internal/game/card"
	"jokenpo/internal/game/deck"
	"log"
	"time"
)

// PlayCardAction é a mensagem interna (vinda da API) para a ação de jogar uma carta.
type PlayCardAction struct {
	PlayerID  string
	CardIndex int
}

// ============================================================================
// Lógica Principal do Jogo (Handlers de Fase Adaptados)
// ============================================================================

// startGame embaralha os decks, compra as mãos iniciais e inicia a primeira rodada.
func (gr *GameRoom) startGame() {
	// --- MODIFICAÇÃO ---
	// A verificação de fase agora usa os métodos get/set thread-safe.
	if gr.getGameState() != phase_ROOM_START {
		gr.handleGameOver("", "Game start failed: invalid phase.")
		return
	}

	drawStatus := make(map[string]bool)

	for playerID, pInfo := range gr.players {
		pInfo.GameDeck.Shuffle(deck.DECK,gr.rng)
		drawStatus[playerID] = gr.drawCardsAndNotify(playerID, initial_HAND_SIZE)
	}

	if gr.checkDeckOutWinCondition(drawStatus) {
		return
	}

	log.Printf("[GameRoom %s] Match started, timer of 5s activated.", gr.ID)

	// --- MODIFICAÇÃO ---
	// Notifica os jogadores via callback HTTP que o jogo começou.
	gr.broadcastEvent("GAME_START", map[string]string{
		"message": "The match has started! You have 5 seconds to play your card.",
	})

	gr.setGameState(phase_WAITING_FOR_PLAYS)
	gr.roundTimer = time.NewTimer(5 * time.Second)
}

// startNewRound compra uma nova carta para cada jogador e inicia a próxima rodada.
func (gr *GameRoom) startNewRound() {
	if gr.getGameState() != phase_ROUND_START {
		gr.handleGameOver("", "Round start failed: invalid phase.")
		return
	}

	gr.playedCards = make(map[string]*card.Card)
	drawStatus := make(map[string]bool)

	for playerID := range gr.players {
		drawStatus[playerID] = gr.drawCardsAndNotify(playerID, 1)
	}

	if gr.checkDeckOutWinCondition(drawStatus) {
		return
	}

	// --- MODIFICAÇÃO ---
	// Notifica os jogadores via callback HTTP.
	gr.broadcastEvent("NEW_ROUND", map[string]string{
		"message": "A new round has started! You have 5 seconds to play your card.",
	})

	gr.setGameState(phase_WAITING_FOR_PLAYS)
	gr.roundTimer = time.NewTimer(5 * time.Second)
}

// HandlePlayCard processa a jogada de um jogador.
func (gr *GameRoom) HandlePlayCard(playerID string, cardIndex int) {
	// --- MODIFICAÇÃO ---
	// Toda a lógica foi adaptada para usar `playerID` e `gr.players`
	// em vez de `*PlayerSession`, e para enviar callbacks em vez de mensagens de WebSocket.
	if gr.getGameState() != phase_WAITING_FOR_PLAYS {
		gr.sendCallbackToPlayer(playerID, "ERROR", map[string]string{"message": "It's not time to play a card right now."})
		return
	}
	if _, alreadyPlayed := gr.playedCards[playerID]; alreadyPlayed {
		gr.sendCallbackToPlayer(playerID, "ERROR", map[string]string{"message": "You have already played a card this round."})
		return
	}

	pInfo := gr.players[playerID]
	if pInfo == nil { return } // Verificação de segurança

	playedCard, err := pInfo.GameDeck.PlayCardFromHand(cardIndex)
	if err != nil {
		gr.sendCallbackToPlayer(playerID, "ERROR", map[string]string{"message": fmt.Sprintf("Failed to play card: %v", err)})
		return
	}

	gr.playedCards[playerID] = playedCard

	gr.sendCallbackToPlayer(playerID, "PLAY_CONFIRMED", map[string]string{
		"message": fmt.Sprintf("You played %s. Waiting for opponent...", playedCard.Key()),
	})
	opponentID := gr.getOpponentID(playerID)
	gr.sendCallbackToPlayer(opponentID, "OPPONENT_PLAYED", map[string]string{
		"message": "Your opponent has played a card.",
	})

	if len(gr.playedCards) == len(gr.players) {
		gr.roundTimer.Stop()
		gr.setGameState(phase_RESOLVING_ROUND)
		// A resolução agora é chamada pela goroutine Run para evitar bloqueio.
	}
}

// resolveRound compara as cartas e determina o resultado da rodada.
func (gr *GameRoom) resolveRound() {
	// --- MODIFICAÇÃO ---
	// Lógica adaptada para usar `playerID` e `pInfo`.
	if gr.getGameState() != phase_RESOLVING_ROUND {
		gr.handleGameOver("", "Round resolution failed: invalid phase.")
		return
	}

	playerIDs := gr.getPlayerIDs()
	if len(playerIDs) != 2 { return }
	p1ID, p2ID := playerIDs[0], playerIDs[1]

	p1Info, p2Info := gr.players[p1ID], gr.players[p2ID]
	p1Card, p2Card := gr.playedCards[p1ID], gr.playedCards[p2ID]

	// Verifica se ambas as cartas foram jogadas antes de comparar
	if p1Card == nil || p2Card == nil {
		gr.handleGameOver("", "Failed to resolve round: one or more players did not play a card.")
		return
	}

	winnerResult := card.Compare(p1Card, p2Card)
	var p1Won, p2Won bool
	var resultText string
	
	switch winnerResult {
	case card.Card1Wins:
		p1Won, p2Won = true, false
		resultText = fmt.Sprintf("Player %s's %s wins against Player %s's %s!", p1ID, p1Card.Key(), p2ID, p2Card.Key())
	case card.Card2Wins:
		p1Won, p2Won = false, true
		resultText = fmt.Sprintf("Player %s's %s wins against Player %s's %s!", p2ID, p2Card.Key(), p1ID, p1Card.Key())
	case card.Tie:
		p1Won, p2Won = false, false
		resultText = fmt.Sprintf("It's a tie between %s and %s!", p1Card.Key(), p2Card.Key())
	}
	
	p1Info.GameDeck.ResolvePlay(p1Won)
	p2Info.GameDeck.ResolvePlay(p2Won)

	gr.broadcastEvent("ROUND_RESULT", map[string]interface{}{
		"message":    resultText,
		"p1_card":    p1Card.Key(),
		"p2_card":    p2Card.Key(),
	})
	
	p1HasWon := p1Info.GameDeck.WinCondition()
	p2HasWon := p2Info.GameDeck.WinCondition()

	if p1HasWon && p2HasWon {
		gr.handleGameOver("", "Both players met win conditions simultaneously.")
		return
	}
	if p1HasWon {
		gr.handleGameOver(p1ID, "Player 1 met the win condition.")
		return
	}
	if p2HasWon {
		gr.handleGameOver(p2ID, "Player 2 met the win condition.")
		return
	}

	time.Sleep(3 * time.Second)
	gr.setGameState(phase_ROUND_START)
	gr.startNewRound()
}

// handleGameOver finaliza a partida e notifica os jogadores.
func (gr *GameRoom) handleGameOver(winnerID string, reason string) {
	// --- MODIFICAÇÃO ---
	// Lógica adaptada para usar `playerID` e `broadcastEvent`.
	if gr.IsFinished() { return }
	gr.setGameState(phase_GAME_OVER)
	
	if gr.roundTimer != nil {
		gr.roundTimer.Stop()
	}
	log.Printf("[GameRoom %s] Game Over. Winner: %s. Reason: %s", gr.ID, winnerID, reason)
	
	gr.broadcastEvent("GAME_OVER", map[string]interface{}{
		"winnerId": winnerID,
		"reason":   reason,
	})

	close(gr.quit) // Sinaliza para a goroutine Run() terminar.
}

// handleTimeout força a jogada de jogadores que não agiram a tempo.
func (gr *GameRoom) handleTimeout() {
	// --- MODIFICAÇÃO ---
	// Lógica adaptada para usar `playerID` e callbacks.
	if gr.getGameState() != phase_WAITING_FOR_PLAYS { return }
	log.Printf("[GameRoom %s] Round timer expired. Forcing remaining plays.", gr.ID)

	gr.setGameState(phase_RESOLVING_ROUND)

	for playerID, pInfo := range gr.players {
		if _, hasPlayed := gr.playedCards[playerID]; !hasPlayed {
			hand, _ := pInfo.GameDeck.GetCardsInZone("hand")
			if len(hand) == 0 {
				opponentID := gr.getOpponentID(playerID)
				gr.handleGameOver(opponentID, fmt.Sprintf("Player %s timed out with no playable cards.", playerID))
				return
			}
			
			playedCard, err := pInfo.GameDeck.PlayRandomCardFromHand(gr.rng)
			if err != nil {
				opponentID := gr.getOpponentID(playerID)
				gr.handleGameOver(opponentID, fmt.Sprintf("Critical error forcing play for %s.", playerID))
				return
			}
			gr.playedCards[playerID] = playedCard

			gr.sendCallbackToPlayer(playerID, "FORCED_PLAY", map[string]string{
				"message": fmt.Sprintf("You ran out of time! The card %s was played for you.", playedCard.Key()),
			})
		}
	}
}


// ============================================================================
// Funções Helper (Adaptadas)
// ============================================================================

// drawCardsAndNotify compra cartas para um jogador e o notifica via callback.
func (gr *GameRoom) drawCardsAndNotify(playerID string, numToDraw int) bool {
	// --- MODIFICAÇÃO ---
	// Lógica completamente reescrita para usar `playerID` e enviar callbacks HTTP.
	pInfo, ok := gr.players[playerID]
	if !ok { return false } // O jogador não está na sala.
	
	drawSuccessful := true
	var warningMessage string

	for i := 0; i < numToDraw; i++ {
		if _, err := pInfo.GameDeck.DrawToHand(); err != nil {
			warningMessage = "Warning: Not enough cards in your deck."
			drawSuccessful = false
			break
		}
	}

	hand, _ := pInfo.GameDeck.GetCardsInZone("hand")
	handKeys := make([]string, len(hand))
	for i, c := range hand {
		handKeys[i] = c.Key()
	}
	
	log.Printf("[GameRoom %s] Player %s hand updated (cards: %d). drawSuccessful=%t", gr.ID, playerID, len(handKeys), drawSuccessful)
	gr.sendCallbackToPlayer(playerID, "UPDATE_HAND", map[string]interface{}{
	"message": warningMessage,
	"hand":    handKeys,
})

	return drawSuccessful
}

// checkDeckOutWinCondition verifica se algum jogador venceu por falta de cartas do oponente.
func (gr *GameRoom) checkDeckOutWinCondition(drawStatus map[string]bool) bool {
	// --- MODIFICAÇÃO ---
	// Adaptado para usar `playerID` em vez de ponteiros.
	playerIDs := gr.getPlayerIDs()
	if len(playerIDs) != 2 { return false }
	p1ID, p2ID := playerIDs[0], playerIDs[1]
	p1DrawOK := drawStatus[p1ID]
	p2DrawOK := drawStatus[p2ID]

	if !p1DrawOK && p2DrawOK {
		gr.handleGameOver(p2ID, "Player 1 ran out of cards.")
		return true
	}
	if !p2DrawOK && p1DrawOK {
		gr.handleGameOver(p1ID, "Player 2 ran out of cards.")
		return true
	}
	if !p1DrawOK && !p2DrawOK {
		gr.handleGameOver("", "Both players ran out of cards.")
		return true
	}
	return false
}

// getOpponentID é o novo utilitário para encontrar o ID do oponente.
func (gr *GameRoom) getOpponentID(playerID string) string {
	for id := range gr.players {
		if id != playerID {
			return id
		}
	}
	return ""
}
//END OF FILE jokenpo/internal/services/gameroom/room_logic.go