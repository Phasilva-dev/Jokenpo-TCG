//START OF FILE jokenpo/internal/services/gameroom/manager.go
package gameroom

import (
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// InitialPlayerInfo é o DTO que vem da API para criar uma sala.
type InitialPlayerInfo struct {
	ID          string   `json:"playerId"`
	CallbackURL string   `json:"callbackUrl"`
	Deck        []string `json:"deck"`
}

// RoomManager (o ator) gerencia o ciclo de vida de todas as salas ativas.
type RoomManager struct {
	rooms      map[string]*GameRoom
	requestCh  chan interface{}
	httpClient *http.Client
}

func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms:      make(map[string]*GameRoom),
		requestCh:  make(chan interface{}),
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// --- Mensagens para o Ator RoomManager ---
type createRoomRequest struct {
	PlayerInfos []*InitialPlayerInfo
	reply       chan *GameRoom
}
type getRoomRequest struct {
	roomID string
	reply  chan *GameRoom
}
type cleanupFinishedRooms struct{}

// --- APIs Públicas do Ator ---

// CreateRoom envia um pedido para o ator para criar uma nova sala.
func (rm *RoomManager) CreateRoom(p1, p2 *InitialPlayerInfo) *GameRoom {
	reply := make(chan *GameRoom)
	rm.requestCh <- createRoomRequest{
		PlayerInfos: []*InitialPlayerInfo{p1, p2},
		reply:       reply,
	}
	return <-reply
}

// GetRoom envia um pedido para o ator para obter uma referência a uma sala existente.
// Este é o método crucial que o handler da API usará para rotear ações.
func (rm *RoomManager) GetRoom(roomID string) *GameRoom {
	reply := make(chan *GameRoom)
	rm.requestCh <- getRoomRequest{roomID: roomID, reply: reply}
	return <-reply
}

// Run inicia o loop principal do ator RoomManager.
func (rm *RoomManager) Run() {
	log.Println("[RoomManager] Actor started.")
	cleanupTicker := time.NewTicker(1 * time.Minute)
	defer cleanupTicker.Stop()

	for {
		select {
		case msg := <-rm.requestCh:
			switch req := msg.(type) {
			case createRoomRequest:
				roomID := uuid.NewString()
				room, err := NewGameRoom(roomID, req.PlayerInfos, rm.httpClient)
				if err != nil {
					log.Printf("ERROR: Failed to create new game room: %v", err)
					req.reply <- nil
					continue
				}
				rm.rooms[roomID] = room
				go room.Run()
				req.reply <- room

			case getRoomRequest:
				// Acessa o mapa de forma segura e retorna a sala.
				req.reply <- rm.rooms[req.roomID]

			case cleanupFinishedRooms:
				for id, room := range rm.rooms {
					if room.IsFinished() {
						delete(rm.rooms, id)
						log.Printf("[RoomManager] Cleaned up finished room %s", id)
					}
				}
			}

		case <-cleanupTicker.C:
			// Envia uma mensagem para si mesmo para executar a limpeza de forma segura.
			rm.requestCh <- cleanupFinishedRooms{}
		}
	}
}
//END OF FILE jokenpo/internal/services/gameroom/manager.go