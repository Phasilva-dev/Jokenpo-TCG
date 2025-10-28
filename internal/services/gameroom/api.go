// START OF FILE jokenpo/internal/services/gameroom/api.go
package gameroom

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

// ============================================================================
// DTOs da API
// ============================================================================

// CreateRoomRequest é o DTO que o cliente (jokenpo-session) envia para criar uma sala.
// Note que usamos a struct InitialPlayerInfo.
type CreateRoomRequest struct {
	PlayerInfos []*InitialPlayerInfo `json:"playerInfos"`
}

// CreateRoomResponse é o DTO que este serviço retorna após criar a sala.
type CreateRoomResponse struct {
	RoomID      string `json:"roomId"`
	ServiceAddr string `json:"serviceAddr"` // Endereço deste GameRoomService
}

// PlayCardRequest é o DTO para a ação de jogar uma carta.
type PlayCardRequest struct {
	PlayerID  string `json:"playerId"`
	CardIndex int    `json:"cardIndex"`
}

// ============================================================================
// Configuração dos Handlers
// ============================================================================

// RegisterHandlers configura todas as rotas da API para o GameRoomService.
func RegisterHandlers(mux *http.ServeMux, roomManager *RoomManager, port int) {
	// --- MUDANÇA CRUCIAL ---
	// Lê o endereço anunciado da mesma variável de ambiente que o registro do Consul.
	advertiseAddr := os.Getenv("SERVICE_ADVERTISED_HOSTNAME")
	if advertiseAddr == "" {
		// Se a variável não estiver definida, o serviço não pode funcionar corretamente.
		// Logamos um erro crítico. A criação de sala retornará um endereço inválido.
		log.Printf("CRITICAL: SERVICE_ADVERTISED_HOSTNAME environment variable is not set!")
		advertiseAddr = "address-not-configured" // Garante que o problema seja visível
	}
	
	// Handler para criar novas salas.
	mux.HandleFunc("/rooms", handleCreateRoom(roomManager, advertiseAddr, port))
	
	// Handler "coringa" para todas as ações em salas existentes (ex: /rooms/{id}/play).
	mux.HandleFunc("/rooms/", handleRoomAction(roomManager))
}

// ============================================================================
// Implementação dos Handlers
// ============================================================================

// handleCreateRoom lida com a requisição POST /rooms.
func handleCreateRoom(rm *RoomManager, advertiseAddr string, port int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		var req CreateRoomRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.PlayerInfos) != 2 {
			http.Error(w, `{"error": "Invalid payload: requires 'playerInfos' array with 2 players"}`, http.StatusBadRequest)
			return
		}
		
		log.Printf("[DEBUG] GameRoom received CreateRoomRequest.")
		log.Printf("[DEBUG] Player 1 (%s) deck size: %d", req.PlayerInfos[0].ID, len(req.PlayerInfos[0].Deck))
		log.Printf("[DEBUG] Player 2 (%s) deck size: %d", req.PlayerInfos[1].ID, len(req.PlayerInfos[1].Deck))

		// Chama o RoomManager para criar a sala de forma síncrona.
		room := rm.CreateRoom(req.PlayerInfos[0], req.PlayerInfos[1])
		if room == nil {
			http.Error(w, `{"error": "Failed to create room"}`, http.StatusInternalServerError)
			return
		}

		// Retorna o ID da sala e o endereço deste serviço para o jokenpo-session.
		resp := CreateRoomResponse{
			RoomID:      room.ID,
			ServiceAddr: fmt.Sprintf("%s:%d", advertiseAddr, port),
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(resp)

		log.Printf("[handleCreateRoom] Response sent for room %s. Now sending start signal.", room.ID)

		// --- CORREÇÃO FINAL ---
		// SÓ DEPOIS de responder, nós damos o sinal para a goroutine do jogo começar.
		room.StartGame()
	}
}

// handleRoomAction é um roteador para ações em salas existentes.
func handleRoomAction(rm *RoomManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extrai o RoomID e a Ação da URL. Ex: /rooms/uuid-123/play -> ["uuid-123", "play"]
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/rooms/"), "/")
		if len(parts) < 1 || parts[0] == "" {
			http.Error(w, `{"error": "Malformed URL, expecting /rooms/{id}/{action}"}`, http.StatusBadRequest)
			return
		}
		roomID := parts[0]

		// Pede ao RoomManager a referência para a sala de forma segura.
		room := rm.GetRoom(roomID)
		if room == nil {
			http.Error(w, `{"error": "Room not found"}`, http.StatusNotFound)
			return
		}

		// Se a URL tiver uma ação (ex: /play), roteia para o handler correto.
		if len(parts) > 1 {
			action := parts[1]
			switch action {
			case "play":
				if r.Method == http.MethodPost {
					handlePlayCardAction(w, r, room)
				} else {
					http.Error(w, `{"error": "Use POST for /play action"}`, http.StatusMethodNotAllowed)
				}
			// Futuramente: case "surrender": ...
			default:
				http.Error(w, `{"error": "Unknown room action"}`, http.StatusNotFound)
			}
		} else {
			// Se for apenas /rooms/{id}, poderia retornar o estado da sala (opcional)
			http.Error(w, `{"error": "Action required, e.g., /play"}`, http.StatusBadRequest)
		}
	}
}

// handlePlayCardAction decodifica o pedido de jogada e o encaminha para a goroutine da sala.
func handlePlayCardAction(w http.ResponseWriter, r *http.Request, room *GameRoom) {
	var req PlayCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid payload for play action"}`, http.StatusBadRequest)
		return
	}

	// Cria a mensagem para a goroutine da sala
	action := PlayCardAction{
		PlayerID:  req.PlayerID,
		CardIndex: req.CardIndex,
	}

	// Envia a ação para o canal 'incoming' da sala correta.
	room.ForwardAction(action)

	w.WriteHeader(http.StatusAccepted) // 202 Accepted: a jogada foi recebida.
}

//END OF FILE jokenpo/internal/services/gameroom/api.go