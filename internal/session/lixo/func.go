package session

/*
func (h *GameHandler) CreateNewRoom(p1, p2 *PlayerSession) {
	// Tenta preparar o primeiro jogador.
	if err := p1.Player.StartPlay(); err != nil {
		p1.Player.EndPlay()
		// Se p1 falhar, p2 é o "inocente".
		h.handleRoomCreationError(p1, p2, err)
		return
	}

	// Tenta preparar o segundo jogador.
	if err := p2.Player.StartPlay(); err != nil {
		p2.Player.EndPlay()
		// Se p2 falhar, p1 é o "inocente".
		h.handleRoomCreationError(p2, p1, err)
		return
	}

	// --- SUCESSO ---
	// Se chegamos aqui, ambos os jogadores estão prontos.

	p1.State = state_IN_MATCH
	p2.State = state_IN_MATCH

	// TODO: Criar e gerenciar a GameRoom.
	roomID := uuid.New().String()

	// 2. Cria a instância da GameRoom, passando o ID e o canal 'finished'.
	room := NewGameRoom(roomID, p1, p2, h.roomFinished)

	// 3. Armazena a sala no mapa de salas ativas do GameHandler.
	h.rooms[roomID] = room
	p1.CurrentRoom = room
	p2.CurrentRoom = room

	// 4. Inicia a goroutine da sala de jogo. É AQUI QUE O JOGO COMEÇA!
	go room.Run()
	
	msg := message.CreateSuccessResponse(state_IN_MATCH,"Match found! The game is starting.", nil)
	p1.Client.Send() <- msg
	p2.Client.Send() <- msg

	fmt.Printf("Game room created successfully for %s and %s.\n", p1.Client.Conn().RemoteAddr(), p2.Client.Conn().RemoteAddr())
}
*/