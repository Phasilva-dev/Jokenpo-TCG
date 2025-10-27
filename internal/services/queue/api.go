//START OF FILE jokenpo/internal/services/queue/api.go
package queue

import (
	"encoding/json"
	"jokenpo/internal/services/cluster"
	"net/http"
)

// ============================================================================
// DTOs (Data Transfer Objects)
// ============================================================================

// EnqueueMatchRequest define o payload para entrar na fila de partida.
type EnqueueMatchRequest struct {
	PlayerID    string   `json:"playerId"`
	CallbackURL string   `json:"callbackUrl"`
	Deck        []string `json:"deck"` // <-- ADICIONE ESTE CAMPO
}

// EnqueueTradeRequest define o payload para entrar na fila de troca às cegas.
type EnqueueTradeRequest struct {
	PlayerID    string `json:"playerId"`
	CallbackURL string `json:"callbackUrl"`
	OfferCard   string `json:"offerCard"`
}

// DequeueRequest define o payload para sair de qualquer fila.
type DequeueRequest struct {
	PlayerID string `json:"playerId"`
}

// ============================================================================
// Configuração dos Handlers
// ============================================================================

// RegisterQueueHandlers configura todas as rotas da API de matchmaking no mux fornecido.
func RegisterQueueHandlers(mux *http.ServeMux, queueMaster *QueueMaster, elector *cluster.LeaderElector) {
	// Cria um "middleware" que envolve nossos handlers para verificar a liderança.
	leaderOnly := leaderOnlyMiddleware(elector)

	// Registra as rotas para match e trade, protegendo-as com o middleware.
	mux.Handle("/queue/match", leaderOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleMatchQueue(w, r, queueMaster)
	})))

	mux.Handle("/queue/trade", leaderOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleTradeQueue(w, r, queueMaster)
	})))
}

// leaderOnlyMiddleware é uma função de ordem superior (higher-order function) que
// retorna um middleware. O middleware verifica a liderança antes de chamar o próximo handler.
func leaderOnlyMiddleware(elector *cluster.LeaderElector) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Se este nó não for o líder, retorna um erro 503 e interrompe.
			if !elector.IsLeader() {
				http.Error(w, `{"error": "This node is not the leader"}`, http.StatusServiceUnavailable)
				return
			}
			// Se for o líder, passa a requisição para o handler principal.
			next.ServeHTTP(w, r)
		})
	}
}

// ============================================================================
// Handlers Específicos das Filas
// ============================================================================

// handleMatchQueue lida com requisições para a fila de partida (POST para entrar, DELETE para sair).
func handleMatchQueue(w http.ResponseWriter, r *http.Request, qm *QueueMaster) {
	switch r.Method {
	case http.MethodPost: // Entrar na fila
		var req EnqueueMatchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "Invalid payload for entering match queue"}`, http.StatusBadRequest)
			return
		}
		player := &PlayerInfo{ID: req.PlayerID, CallbackURL: req.CallbackURL}
		qm.EnqueueMatch(player)
		w.WriteHeader(http.StatusAccepted) // 202 Accepted: O pedido foi aceito para processamento futuro.

	case http.MethodDelete: // Sair da fila
		var req DequeueRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "Invalid payload for leaving queue"}`, http.StatusBadRequest)
			return
		}
		qm.DequeueMatch(req.PlayerID)
		w.WriteHeader(http.StatusOK)

	default:
		http.Error(w, `{"error": "Method not allowed. Use POST to enter and DELETE to leave."}`, http.StatusMethodNotAllowed)
	}
}

// handleTradeQueue lida com requisições para a fila de troca.
func handleTradeQueue(w http.ResponseWriter, r *http.Request, qm *QueueMaster) {
	switch r.Method {
	case http.MethodPost: // Entrar na fila
		var req EnqueueTradeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "Invalid payload for entering trade queue"}`, http.StatusBadRequest)
			return
		}
		trade := &TradeInfo{
			PlayerInfo: PlayerInfo{ID: req.PlayerID, CallbackURL: req.CallbackURL},
			OfferCard:  req.OfferCard,
		}
		qm.EnqueueTrade(trade)
		w.WriteHeader(http.StatusAccepted)

	case http.MethodDelete: // Sair da fila
		var req DequeueRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "Invalid payload for leaving queue"}`, http.StatusBadRequest)
			return
		}
		qm.DequeueTrade(req.PlayerID)
		w.WriteHeader(http.StatusOK)

	default:
		http.Error(w, `{"error": "Method not allowed. Use POST to enter and DELETE to leave."}`, http.StatusMethodNotAllowed)
	}
}

//END OF FILE jokenpo/internal/services/queue/api.go