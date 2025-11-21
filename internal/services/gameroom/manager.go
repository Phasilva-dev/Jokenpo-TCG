//START OF FILE jokenpo/internal/services/gameroom/manager.go
package gameroom

import (
	"jokenpo/internal/services/blockchain" // Importar
	"jokenpo/internal/services/cluster"    // Importar
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
	blockchain *blockchain.BlockchainClient // Novo campo
}

// NewRoomManager agora recebe o ConsulManager para localizar o contrato
func NewRoomManager(manager *cluster.ConsulManager) *RoomManager {
	var bcClient *blockchain.BlockchainClient
	var contractAddr string

	// --- LÓGICA DE ESPERA (POLLING) ---
	// Igual ao Session e Shop: espera o Deployer salvar o endereço
	log.Println("GAMEROOM: Aguardando endereço do contrato no Consul...")
	client := manager.GetClient()

	for i := 0; i < 30; i++ {
		if client == nil {
			client = manager.GetClient()
		}
		if client != nil {
			pair, _, err := client.KV().Get("jokenpo/config/contract_address", nil)
			if err == nil && pair != nil {
				contractAddr = string(pair.Value)
				break
			}
		}
		time.Sleep(2 * time.Second)
	}

	if contractAddr != "" {
		var err error
		// Conecta no contrato existente
		bcClient, _, err = blockchain.InitBlockchain(contractAddr)
		if err != nil {
			log.Printf("GAMEROOM AVISO: Erro ao conectar na blockchain: %v", err)
		} else {
			log.Printf("GAMEROOM: Conectado à Blockchain em %s", contractAddr)
		}
	} else {
		log.Println("GAMEROOM AVISO: Timeout aguardando contrato. Auditoria desabilitada.")
	}

	return &RoomManager{
		rooms:      make(map[string]*GameRoom),
		requestCh:  make(chan interface{}),
		httpClient: &http.Client{Timeout: 10 * time.Second},
		blockchain: bcClient, // Armazena o cliente
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

func (rm *RoomManager) CreateRoom(p1, p2 *InitialPlayerInfo) *GameRoom {
	reply := make(chan *GameRoom)
	rm.requestCh <- createRoomRequest{
		PlayerInfos: []*InitialPlayerInfo{p1, p2},
		reply:       reply,
	}
	return <-reply
}

func (rm *RoomManager) GetRoom(roomID string) *GameRoom {
	reply := make(chan *GameRoom)
	rm.requestCh <- getRoomRequest{roomID: roomID, reply: reply}
	return <-reply
}

// --- Helper ---
func (rm *RoomManager) handleMessage(msg interface{}) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("CRITICAL: Recovered from panic in RoomManager: %v", r)
		}
	}()

	switch req := msg.(type) {
	case createRoomRequest:
		roomID := uuid.NewString()
		// CORREÇÃO DO ERRO: Agora passamos rm.blockchain como 4º argumento
		room, err := NewGameRoom(roomID, req.PlayerInfos, rm.httpClient, rm.blockchain)
		
		log.Printf("[DEBUG] Created Room %s", roomID)
		if err != nil {
			log.Printf("ERROR: Failed to create new game room: %v", err)
			req.reply <- nil
			return
		}
		rm.rooms[roomID] = room
		go room.Run()
		req.reply <- room

	case getRoomRequest:
		req.reply <- rm.rooms[req.roomID]

	case cleanupFinishedRooms:
		for id, room := range rm.rooms {
			if room.IsFinished() {
				delete(rm.rooms, id)
				log.Printf("[RoomManager] Cleaned up finished room %s", id)
			}
		}
	}
}

func (rm *RoomManager) Run() {
	log.Println("[RoomManager] Actor started.")
	cleanupTicker := time.NewTicker(1 * time.Minute)
	defer cleanupTicker.Stop()

	for {
		select {
		case msg := <-rm.requestCh:
			rm.handleMessage(msg)

		case <-cleanupTicker.C:
			rm.handleMessage(cleanupFinishedRooms{})
		}
	}
}
//END OF FILE jokenpo/internal/services/gameroom/manager.go