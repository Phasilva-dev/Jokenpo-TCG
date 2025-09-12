package session

import (
	"fmt"
	"jokenpo/internal/game/card"
	"jokenpo/internal/network"
	"jokenpo/internal/session/message"
	"strings"
)

// --------- Funções Privadas ---------

// broadcast é uma função de conveniência para enviar a mesma mensagem para ambos os jogadores.
func (gr *GameRoom) broadcast(msg network.Message) {
	for _, p := range gr.players {
		p.Client.Send() <- msg
	}
}

// Compra as cartas de um jogador, notifica ele e retorna um True (Conseguimos comprar) ou False (Acabou as cartas, deck out)
func (gr *GameRoom) drawCardsAndNotify(p *PlayerSession, numToDraw int) bool {
	var warningMessage string
	drawSuccessful := true // Começamos assumindo que a compra será bem-sucedida.

	var card *card.Card
	// 1. Tenta comprar o número de cartas especificado.
	for i := 0; i < numToDraw; i++ {
		var err error
		card, err = p.Player.DrawToHand()
		if err != nil {
			warningMessage = "Warning: Not enough cards in your deck. Play with the cards you received. \n"
			drawSuccessful = false
			break
		}
		p.Client.Send()
	}

	// 2. Após as tentativas de compra, pega a visão da mão atual do jogador.
	handStr, err := p.Player.SeeHand()
	if err != nil {
		// Erro critico, encerra o game em empate
		msg := fmt.Sprintf("critical error in the room %s: failed to see player's hand: %v\n", gr.ID, err)
		gr.handleGameOver(nil, msg)
	}

	// 3. Monta a mensagem final e personalizada para este jogador.
	var finalMsgBuilder strings.Builder

	// Adiciona o aviso, se houver um.
	if warningMessage != "" {
		finalMsgBuilder.WriteString("\n" + warningMessage)
	} else {
		if numToDraw == 1 && card != nil {
			msg := fmt.Sprintf("You drew this card: %s \n", card.String())
			finalMsgBuilder.WriteString(msg)
		}
	}

	// Anexa o estado atual da mão.
	finalMsgBuilder.WriteString("Your current hand:\n")
	finalMsgBuilder.WriteString(handStr)
	finalMsgBuilder.WriteString("\n Escolha o indice da sua carta: ")

	// 4. Cria e envia a mensagem para o jogador.
	msg := message.CreateSuccessResponse(finalMsgBuilder.String(), nil)
	p.Client.Send() <- msg

	return drawSuccessful
	//Função já mostra a mão e seus indices
}

//Função para checkar se alguém ganhou por deckout
func (gr *GameRoom) checkDeckOutWinCondition(drawStatus map[*PlayerSession]bool) bool {
	
	p1 := gr.players[0]
	p2 := gr.players[1]
	p1DrawOK := drawStatus[p1]
	p2DrawOK := drawStatus[p2]

	// Cenário 1: P1 não conseguiu comprar, mas P2 sim. P2 vence.
	if !p1DrawOK && p2DrawOK {
		gr.handleGameOver(p2, "Player 1 ran out of cards.")
		return true // O jogo terminou.
	}

	// Cenário 2: P2 não conseguiu comprar, mas P1 sim. P1 vence.
	if !p2DrawOK && p1DrawOK {
		gr.handleGameOver(p1, "Player 2 ran out of cards.")
		return true // O jogo terminou.
	}

	// Cenário 3: Ambos não conseguiram comprar. É um empate.
	if !p1DrawOK && !p2DrawOK {
		gr.handleGameOver(nil, "Both players ran out of cards.") // 'nil' indica empate.
		return true // O jogo terminou.
	}

	return false // Nenhuma condição de Deck Out foi encontrada. O jogo continua.
}

// getOpponent é um pequeno utilitário para encontrar o oponente de um jogador na sala.
func (gr *GameRoom) getOpponent(p *PlayerSession) *PlayerSession {
	if gr.players[0] == p {
		return gr.players[1]
	}
	return gr.players[0]
}

func (gr *GameRoom) closeRoom() {
	// 1. Sinaliza para a própria goroutine (Run) que ela deve parar.
	// Usamos close() em vez de enviar um valor. É um padrão comum para canais de sinalização.
	close(gr.quit)

	// 2. Notifica o GameHandler que esta sala terminou, enviando seu ID.
	gr.finished <- gr.ID
}