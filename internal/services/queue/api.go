// START OF FILE jokenpo/internal/services/queue/api.go
package queue

import (
	"encoding/json"
	"jokenpo/internal/services/cluster"
	"log"
	"net/http"
)

// ============================================================================
// DTOs (Data Transfer Objects)
// ============================================================================

type EnqueueMatchRequest struct {
	PlayerID    string   `json:"playerId"`
	CallbackURL string   `json:"callbackUrl"` // Esta será a URL para /game-event
	Deck        []string `json:"deck"`
}

type EnqueueTradeRequest struct {
	PlayerID    string `json:"playerId"`
	CallbackURL string `json:"callbackUrl"` // Esta será a URL para /trade-found
	OfferCard   string `json:"offerCard"`
}

type DequeueRequest struct {
	PlayerID string `json:"playerId"`
}

// ============================================================================
// Configuração dos Handlers
// ============================================================================

func RegisterQueueHandlers(mux *http.ServeMux, queueMaster *QueueMaster, elector *cluster.LeaderElector) {
	leaderOnly := leaderOnlyMiddleware(elector)
	mux.Handle("/queue/match", leaderOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleMatchQueue(w, r, queueMaster)
	})))
	mux.Handle("/queue/trade", leaderOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleTradeQueue(w, r, queueMaster)
	})))
}

func leaderOnlyMiddleware(elector *cluster.LeaderElector) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !elector.IsLeader() {
				http.Error(w, `{"error": "This node is not the leader"}`, http.StatusServiceUnavailable)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ============================================================================
// Handlers Específicos das Filas
// ============================================================================

func handleMatchQueue(w http.ResponseWriter, r *http.Request, qm *QueueMaster) {
	switch r.Method {
	case http.MethodPost:
		var req EnqueueMatchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "Invalid payload for entering match queue"}`, http.StatusBadRequest)
			return
		}
		
		// O callback para o resultado do match vem do query param da URL.
		matchCallbackURL := r.URL.Query().Get("callback")
		if matchCallbackURL == "" {
			http.Error(w, `{"error": "Missing 'callback' query parameter"}`, http.StatusBadRequest)
			return
		}

		log.Printf("[DEBUG] Queue received EnqueueMatchRequest for Player %s. GameCallback: %s, MatchCallback: %s", req.PlayerID, req.CallbackURL, matchCallbackURL)

		player := &PlayerInfo{
			ID:               req.PlayerID,
			CallbackURL:      req.CallbackURL, // A URL para /game-event que será passada ao GameRoom
			MatchCallbackURL: matchCallbackURL,  // A URL para /match-found que o Queue usará
			Deck:             req.Deck,
		}
		qm.EnqueueMatch(player)
		w.WriteHeader(http.StatusAccepted)

	case http.MethodDelete:
		var req DequeueRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "Invalid payload for leaving queue"}`, http.StatusBadRequest)
			return
		}
		qm.DequeueMatch(req.PlayerID)
		w.WriteHeader(http.StatusOK)

	default:
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func handleTradeQueue(w http.ResponseWriter, r *http.Request, qm *QueueMaster) {
	switch r.Method {
	case http.MethodPost:
		var req EnqueueTradeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "Invalid payload for entering trade queue"}`, http.StatusBadRequest)
			return
		}
		trade := &TradeInfo{
			PlayerInfo: PlayerInfo{
				ID:          req.PlayerID,
				CallbackURL: req.CallbackURL, // Para trocas, esta é a única URL necessária.
			},
			OfferCard: req.OfferCard,
		}
		qm.EnqueueTrade(trade)
		w.WriteHeader(http.StatusAccepted)

	case http.MethodDelete:
		var req DequeueRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "Invalid payload for leaving queue"}`, http.StatusBadRequest)
			return
		}
		qm.DequeueTrade(req.PlayerID)
		w.WriteHeader(http.StatusOK)

	default:
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
	}
}
//END OF FILE jokenpo/internal/services/queue/api.go