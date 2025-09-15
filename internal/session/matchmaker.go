package session

import (
	"fmt"
	"jokenpo/internal/session/message"
	"strings"
	"time"
)

type Matchmaker struct {
	queue []*PlayerSession

	// Canal para adicionar jogadores à fila de forma segura e concorrente.
	enqueue chan *PlayerSession

	dequeue chan *PlayerSession

	// Uma referência de volta ao GameHandler para que o Matchmaker possa
	// criar salas de jogo e atualizar o estado dos jogadores.
	gameHandler *GameHandler
}

// NewMatchmaker cria e inicializa um novo Matchmaker.
func NewMatchmaker(gh *GameHandler) *Matchmaker {
	return &Matchmaker{
		queue:       make([]*PlayerSession, 0),
		enqueue:     make(chan *PlayerSession),
		dequeue:     make(chan *PlayerSession),
		gameHandler: gh,
	}
}

// Run inicia o loop do Matchmaker em sua própria goroutine.
func (m *Matchmaker) Run() {
	fmt.Println("Matchmaker started.")
	// O ticker vai "disparar" a cada segundo, nos dando uma chance de verificar a fila.
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		
		// Um novo jogador quer entrar na fila.
		case playerSession := <-m.enqueue:
			m.queue = append(m.queue, playerSession)
			fmt.Printf("Player added to queue matchmaking. queue now have %d players.\n", len(m.queue))
			
			// Informa ao jogador que ele está na fila.
			msg := message.CreateSuccessResponse(playerSession.State, "You are now in the matchmaking queue.", nil)
			playerSession.Client.Send() <- msg


		case playerToLeave := <- m.dequeue:
			for i, playerInQueue := range m.queue {
				if playerInQueue == playerToLeave {
					// Encontramos! Removemos o jogador usando o truque de slice do Go.
					m.queue = append(m.queue[:i], m.queue[i+1:]...)
					
					// Log no servidor para depuração.
					fmt.Printf("Player %s left matchmaking queue. Queue size now: %d\n",
						playerToLeave.Client.Conn().RemoteAddr(), len(m.queue))

					// Envia uma mensagem de confirmação para o jogador.
					//msg := message.CreateSuccessResponse("You have left the matchmaking queue.", nil)
					//playerToLeave.Client.Send() <- msg
					
					// Para o loop 'for' pois já encontramos e removemos o jogador.
					break 
				}
			}
		// O ticker disparou, hora de verificar se podemos formar um par.
		case <-ticker.C:
			if len(m.queue) >= 2 {
				// Temos um par!
				player1 := m.queue[0]
				player2 := m.queue[1]
				
				// Remove os dois da frente da fila.
				m.queue = m.queue[2:]
				
				fmt.Printf("Match found! %s vs %s. in queue now: %d\n", player1.Client.Conn().RemoteAddr(), player2.Client.Conn().RemoteAddr(), len(m.queue))
				
				// Delega a criação da sala de jogo para o GameHandler.
				m.gameHandler.CreateNewRoom(player1, player2)
			} else {
				m.broadcastQueue()
			}
		}
	}
}

// EnqueuePlayer adiciona um jogador à fila de matchmaking de forma segura.
func (m *Matchmaker) EnqueuePlayer(session *PlayerSession) {
	// Apenas envia a sessão para o canal. A goroutine Run do Matchmaker
	// fará o resto do trabalho. Isso é rápido e não bloqueia.
	m.enqueue <- session
}

func (m *Matchmaker) broadcastQueue() {
	// --- NOVA LÓGICA AQUI ---
	// Se não temos um par, mas há jogadores na fila, vamos atualizá-los.
	// O 'for range' em um slice vazio não faz nada, então é seguro.
	for i, playerInQueue := range m.queue {
	// O índice 'i' é 0-based, então a posição é 'i + 1'.
		position := i + 1

		statusMsg := fmt.Sprintf("Still searching for a match... You are position %d in queue.\n", position)

		var sb strings.Builder

		sb.WriteString(statusMsg)
					
		msg := message.CreateSuccessResponse(playerInQueue.State, sb.String(), nil)
		playerInQueue.Client.Send() <- msg
		playerInQueue.Client.Send() <- message.CreatePromptInputMessage()
	}
}

func (m *Matchmaker) LeaveQueue(session *PlayerSession) {
	m.dequeue <- session
}