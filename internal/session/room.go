package session

import (
	"fmt"
	"jokenpo/internal/game/card"
	"jokenpo/internal/network"
	"jokenpo/internal/session/message"
	"time"
)

const (
	phase_WAITING_FOR_PLAYS = "waiting_for_plays" // A sala está esperando os jogadores fazerem suas jogadas.
	phase_RESOLVING_ROUND  = "resolving_round"  // As jogadas foram feitas, a sala está calculando o resultado.
	phase_GAME_OVER        = "game_over"        // A partida terminou.
	phase_ROOM_START = "room_start" // A partida foi iniciada
	phase_ROUND_START = "round_start"

	initial_HAND_SIZE = 5

	startHeader = "The round has started! You have 30 seconds to play your card.\n\n"
)

type GameRoom struct {
	ID      string
	players []*PlayerSession

	//incoming   chan clientMessage
	unregister chan *PlayerSession

	finished chan<- string     // Canal para notificar o GameHandler que a sala terminou.
	quit     chan struct{}     // Canal para sinalizar o fim da goroutine interna da sala.

	// --- Estado do Jogo ---
	gameState   string                        // Usará as constantes de fase.
	playedCards map[*PlayerSession]*card.Card // Armazena as cartas jogadas na rodada.
	roundTimer  *time.Timer                   // Renomeado de turnTimer para roundTimer.

	// --- Novos campos para o ciclo de vida ---
	
}

func NewGameRoom(id string, p1 *PlayerSession, p2 *PlayerSession, finished chan<- string) *GameRoom {
	return &GameRoom{
		// --- Parâmetros de Entrada ---
		ID: id,
		players: []*PlayerSession{p1, p2},

		// --- Inicialização de Campos Internos ---
		
		unregister: make(chan *PlayerSession),

		finished: finished,

		quit:     make(chan struct{}),
		
		playedCards: make(map[*PlayerSession]*card.Card),

		// --- Estado Inicial do Jogo ---
		gameState: phase_ROOM_START,
		
		roundTimer: nil,
	}
}
// broadcast é uma função de conveniência para enviar a mesma mensagem para ambos os jogadores.
func (gr *GameRoom) broadcast(msg network.Message) {
	for _, p := range gr.players {
		p.Client.Send() <- msg
	}
}

func (gr *GameRoom) startGame() {
	if gr.gameState != phase_ROOM_START {
		return
	}

	gr.playedCards = make(map[*PlayerSession]*card.Card)

	drawStatus := make(map[*PlayerSession]bool)
	for _, p := range gr.players {
		drawStatus[p] = gr.drawCardsAndNotify(p, 1)
	}

	if gr.checkDeckOutWinCondition(drawStatus) {
		return // Se o jogo terminou, não inicie o timer da rodada.
	}

	
	fmt.Printf("Room %s: Match started, timer of 30s activated.\n", gr.ID)
	msg := message.CreateSuccessResponse(startHeader, nil)
	gr.gameState = phase_WAITING_FOR_PLAYS
	gr.roundTimer = time.NewTimer(30 * time.Second)
	gr.broadcast(msg)
}

func (gr *GameRoom) startNewRound() {
	if gr.gameState != phase_ROUND_START {
		return
	}
	gr.playedCards = make(map[*PlayerSession]*card.Card)
	
	drawStatus := make(map[*PlayerSession]bool)
	for _, p := range gr.players {
		drawStatus[p] = gr.drawCardsAndNotify(p, 1)
	}

	if gr.checkDeckOutWinCondition(drawStatus) {
		return // Se o jogo terminou, não inicie o timer da rodada.
	}

	msg := message.CreateSuccessResponse(startHeader, nil)
	gr.gameState = phase_WAITING_FOR_PLAYS
	gr.roundTimer = time.NewTimer(30 * time.Second)
	gr.broadcast(msg)
}

func (gr *GameRoom) HandlePlayCard(player *PlayerSession, cardIndex int) {

	if gr.gameState != phase_WAITING_FOR_PLAYS {
		msg := message.CreateErrorResponse("It's not time to play a card right now.")
		player.Client.Send() <- msg
		return
	}

	// 2. Este jogador já fez uma jogada nesta rodada?
	if _, alreadyPlayed := gr.playedCards[player]; alreadyPlayed {
		msg := message.CreateErrorResponse("You have already played a card this round.")
		player.Client.Send() <- msg
		return
	}

	// 3. Tenta jogar a carta usando a lógica interna do jogador.
	// Isso move a carta da "mão" para a zona de "jogo" e valida o índice.
	playedCard, err := player.Player.PlayCardFromHand(cardIndex)
	if err != nil {
		// O erro mais comum aqui é "índice inválido".
		msg := message.CreateErrorResponse(fmt.Sprintf("Failed to play card: %v", err))
		player.Client.Send() <- msg
		return
	}

	// --- REGISTRO E FEEDBACK ---

	// 4. Registra a carta jogada na sala.
	gr.playedCards[player] = playedCard

	// 5. Envia uma confirmação para o jogador.
	confirmMsg := message.CreateSuccessResponse(fmt.Sprintf("You played %s. Waiting for opponent...", playedCard.Key()), nil)
	player.Client.Send() <- confirmMsg

	// --- VERIFICAÇÃO DE FIM DE RODADA ---

	// 6. Todos os jogadores já jogaram? (Neste caso, 2 jogadores)
	if len(gr.playedCards) == len(gr.players) {
		// Sim! É hora de resolver a rodada.
		gr.roundTimer.Stop() // Para o timer para que ele não dispare desnecessariamente.
		gr.gameState = phase_RESOLVING_ROUND
		gr.resolveRound()
	}
}

func (gr *GameRoom) resolveRound() {
	if gr.gameState != phase_RESOLVING_ROUND {
		return
	}

	// Pega os dois jogadores e suas cartas.
	p1 := gr.players[0]
	p2 := gr.players[1]
	p1Card := gr.playedCards[p1]
	p2Card := gr.playedCards[p2]

	winnerResult := card.Compare(p1Card, p2Card)

	var p1Won, p2Won bool
	var resultText string

	switch winnerResult {
	case card.Card1Wins:
		p1Won, p2Won = true, false
		// Mensagem mais detalhada mostrando tipo e valor.
		resultText = fmt.Sprintf("Player 1's %s wins against Player 2's %s!", p1Card.String(), p2Card.String())
	
	case card.Card2Wins:
		p1Won, p2Won = false, true
		resultText = fmt.Sprintf("Player 2's %s wins against Player 1's %s!", p2Card.String(), p1Card.String())
	
	case card.Tie:
		p1Won, p2Won = false, false
		resultText = fmt.Sprintf("It's a tie between %s and %s!", p1Card.String(), p2Card.String())
	}
	
	// Atualiza o estado interno de cada jogador (move a carta para 'win' ou 'out').
	p1.Player.ResolvePlay(p1Won)
	p2.Player.ResolvePlay(p2Won)

	// Notifica ambos os jogadores sobre o resultado da rodada.
	broadcastMsg := message.CreateSuccessResponse(resultText, nil)
	gr.broadcast(broadcastMsg)

	p1HasWon := p1.Player.Inventory().GameDeck().WinCondition()
	p2HasWon := p2.Player.Inventory().GameDeck().WinCondition()

	// Cenário 1: Ambos os jogadores atingem a condição de vitória na mesma rodada. É um empate.
	if p1HasWon && p2HasWon {
		gr.handleGameOver(nil, "Both players met win conditions simultaneously.")
		return // O jogo acabou.
	}

	// Cenário 2: Apenas P1 venceu.
	if p1HasWon {
		gr.handleGameOver(p1, "Player 1 met the win condition.")
		return // O jogo acabou.
	}

	// Cenário 3: Apenas P2 venceu.
	if p2HasWon {
		gr.handleGameOver(p2, "Player 2 met the win condition.")
		return // O jogo acabou.
	}

	// Cenário 4: Ninguém venceu. O jogo continua para a próxima rodada.
	// É uma boa prática dar um pequeno tempo para os jogadores lerem o resultado.
	time.Sleep(4 * time.Second) // Pausa de 4 segundos.
	
	gr.startNewRound()
}

func (gr *GameRoom) handleGameOver(winner *PlayerSession, reason string) {

	gr.gameState = phase_GAME_OVER

	// 2. Para o timer, caso ele ainda esteja rodando por algum motivo.
	if gr.roundTimer != nil {
		gr.roundTimer.Stop()
	}

	// 3. Loga o fim do jogo no servidor para depuração.
	fmt.Printf("Room %s: Game Over. Reason: %s\n", gr.ID, reason)

	// 4. Lida com os diferentes cenários: Empate ou Vitória de um jogador.
	if winner == nil {
		// --- CENÁRIO DE EMPATE ---
		finalMessage := fmt.Sprintf("Game Over! It's a draw. Reason: %s", reason)
		msg := message.CreateSuccessResponse(finalMessage, nil)
		gr.broadcast(msg)

	} else {
		// --- CENÁRIO DE VITÓRIA ---
		var loser *PlayerSession

		// Encontra o perdedor (o jogador que não é o vencedor).
		if gr.players[0] == winner {
			loser = gr.players[1]
		} else {
			loser = gr.players[0]
		}

		// Cria as mensagens personalizadas.
		winMessageStr := fmt.Sprintf("You Win! Congratulations! Reason: %s", reason)
		loseMessageStr := fmt.Sprintf("You Lose. Better luck next time. Reason: %s", reason)

		// Cria as mensagens de rede.
		winMsg := message.CreateSuccessResponse(winMessageStr, nil)
		loseMsg := message.CreateSuccessResponse(loseMessageStr, nil)
		
		// Envia a mensagem correta para cada jogador.
		winner.Client.Send() <- winMsg
		loser.Client.Send() <- loseMsg
	}

	time.Sleep(3 * time.Second)

	// Mensagem padrão de retorno ao lobby.
	lobbyMessage := "You have returned to the lobby. You can find a new match now."
	msg := message.CreateSuccessResponse(lobbyMessage, nil)

	for _, p := range gr.players {
		// 1. Muda o estado da sessão de volta para o lobby.
		p.State = state_LOBBY
		p.Player.EndPlay()
		
		// 2. Remove a referência para esta sala, limpando o estado do jogador.
		p.CurrentRoom = nil
		
		// 3. Notifica o cliente que ele está de volta ao lobby.
		p.Client.Send() <- msg
	}

	time.Sleep(3 * time.Second)

	gr.closeRoom()
}

/*
// handleTimeout é acionado quando o timer de 30s acaba.
func (gr *GameRoom) handleTimeout() {
	if gr.gameState != phase_WAITING_FOR_PLAYS {
		return
	}
	gr.gameState = phase_RESOLVING_ROUND

	gr.broadcast(message.CreateSuccessResponse("Time is up!", nil))

	// Para cada jogador que ainda não jogou, faz uma jogada aleatória.
	for _, p := range gr.players {
		if _, hasPlayed := gr.playedCards[p]; !hasPlayed {
			card, err := p.Player.PlayRandomCardFromHand() // Você precisará criar este método no Player
			if err == nil {
				gr.playedCards[p] = card
				p.Client.Send() <- message.CreateSuccessResponse(fmt.Sprintf("A random card (%s) was played for you.", card), nil)
			}
		}
	}
	gr.resolveRound()
}


// handleDisconnect trata a desconexão de um jogador no meio da partida.
func (gr *GameRoom) handleDisconnect(disconnectedPlayer *PlayerSession) {
	if gr.gameState == phase_GAME_OVER {
		return // O jogo já acabou, não faz nada.
	}
	
	// O outro jogador é declarado vencedor por W.O.
	var winner *PlayerSession
	for _, p := range gr.players {
		if p != disconnectedPlayer {
			winner = p
			break
		}
	}
	
	gr.broadcast(message.CreateErrorResponse(fmt.Sprintf("Player %s has disconnected.", disconnectedPlayer.Client.Conn().RemoteAddr())))
	gr.endGame(winner)
}*/
