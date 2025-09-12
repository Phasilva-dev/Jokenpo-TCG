package session

import (
	"fmt"
	"jokenpo/internal/game/card"
	"time"
)

const (
	phase_ROOM_START = "room_start" // A partida foi iniciada
	phase_WAITING_FOR_PLAYS = "waiting_for_plays" // A sala está esperando os jogadores fazerem suas jogadas.
	phase_RESOLVING_ROUND  = "resolving_round"  // As jogadas foram feitas, a sala está calculando o resultado.
	phase_GAME_OVER        = "game_over"        // A partida terminou.
	phase_ROUND_START = "round_start"

	initial_HAND_SIZE = 5

	startHeader = "The round has started! You have 30 seconds to play your card.\n\n"
)

type GameRoom struct {
	ID      string
	players []*PlayerSession

	// --- Canais ---
	incoming   chan interface{} // Canal para registrar jogadas
	unregister chan *PlayerSession //Canal para notificar que jogador se desconectou
	finished chan<- string     // Canal para notificar o GameHandler que a sala terminou.
	quit     chan struct{}     // Canal para sinalizar o fim da goroutine interna da sala.

	// --- Estado do Jogo ---
	gameState   string                        // Usará as constantes de fase.
	playedCards map[*PlayerSession]*card.Card // Armazena as cartas jogadas na rodada.
	roundTimer  *time.Timer                   
	
}

func NewGameRoom(id string, p1 *PlayerSession, p2 *PlayerSession, finished chan<- string) *GameRoom {
	return &GameRoom{
		// --- Parâmetros de Entrada ---
		ID: id,
		players: []*PlayerSession{p1, p2},

		// --- Canais ---

		incoming:   make(chan interface{}),
		unregister: make(chan *PlayerSession),
		finished: finished,
		quit:     make(chan struct{}),

		// --- Inicialização de Campos Internos ---
		
		playedCards: make(map[*PlayerSession]*card.Card),
		gameState: phase_ROOM_START,
		roundTimer: nil,
	}
	
}




func (gr *GameRoom) Run() {
	fmt.Printf("Room %s: Goroutine starting. Players: %s, %s\n",
		gr.ID, gr.players[0].Client.Conn().RemoteAddr(), gr.players[1].Client.Conn().RemoteAddr())

	// O defer garante que a sala sempre será limpa e fechada, não importa como o loop termine.
	defer func() {
		if gr.roundTimer != nil {
			gr.roundTimer.Stop()
		}
		gr.closeRoom()
		fmt.Printf("Room %s: Goroutine stopped and cleaned up.\n", gr.ID)
	}()

	// Inicia a partida, compra as cartas iniciais e começa o timer da primeira rodada.
	gr.startGame()

	for {
		select {
		// --- Evento: Ação de um jogador ---
		case action := <- gr.incoming:
			switch act := action.(type) {
			case PlayCardAction:
				gr.HandlePlayCard(act.Session, act.CardIndex)
			}

		// --- Evento: O timer da rodada expirou ---
		case <-gr.roundTimer.C:
			gr.handleTimeout() // Força as jogadas restantes.
			gr.resolveRound()  // Resolve a rodada agora que todas as cartas estão na mesa.

		// --- Evento: Um jogador desconectou ---
		case disconnectedSession := <-gr.unregister:
			// Se o jogo já acabou por outro motivo, não fazemos nada.
			if gr.gameState == phase_GAME_OVER {
				return // A limpeza já está em andamento.
			}
			// O oponente vence.
			winner := gr.getOpponent(disconnectedSession)
			reason := fmt.Sprintf("Opponent %s disconnected.", disconnectedSession.Client.Conn().RemoteAddr())
			
			// handleGameOver cuida de tudo: notificar, limpar estados e chamar closeRoom.
			gr.handleGameOver(winner, reason)
			
			// Como o jogo acabou, podemos sair do loop imediatamente.
			return

		// --- Evento: A sala recebeu um sinal interno para fechar ---
		case <-gr.quit:
			// Este sinal vem de `closeRoom()`, chamado por `handleGameOver`.
			// Apenas retornamos para que o 'defer' possa fazer a limpeza final.
			return
		}
	}
	
}
