//START OF FILE jokenpo/internal/services/shop/api.go
package shop

import (
	"encoding/json"
	"jokenpo/internal/services/cluster"
	"log"
	"net/http"
)

// DTOs (Data Transfer Objects) para o contrato da API do Shop
type PurchaseRequest struct {
	PlayerID string `json:"playerId"` // Campo Obrigatório
	Quantity uint64 `json:"quantity"`
}

type PurchaseResponse struct {
	Cards []string `json:"cards,omitempty"`
	Error string   `json:"error,omitempty"`
}

// CreateShopHandler cria o handler HTTP para o ShopService.
// Ele garante que apenas o líder processe as requisições e que o estado
// seja persistido antes de confirmar a operação para o cliente.
func CreateShopHandler(shopService *ShopService, elector *cluster.LeaderElector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. VERIFICAÇÃO DE LIDERANÇA
		if !elector.IsLeader() {
			http.Error(w, `{"error": "This node is not the leader and cannot process write operations"}`, http.StatusServiceUnavailable)
			return
		}

		// 2. DECODIFICA O REQUEST
		var req PurchaseRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "Invalid payload"}`, http.StatusBadRequest)
			return
		}

		if req.PlayerID == "" {
			http.Error(w, `{"error": "playerId is required"}`, http.StatusBadRequest)
			return
		}

		// 3. EXECUTA A LÓGICA DE NEGÓCIO (EM MEMÓRIA)
		// Passamos o PlayerID para o serviço para registro na blockchain
		cards, err := shopService.Purchase(req.PlayerID, req.Quantity)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(PurchaseResponse{Error: err.Error()})
			return
		}

		// 4. PERSISTE O ESTADO (WRITE-AHEAD) - CONSUL
		if err := elector.PersistState(shopService); err != nil {
			log.Printf("CRITICAL: State changed in memory but failed to persist to Consul: %v. Transaction will NOT be confirmed.", err)
			http.Error(w, `{"error": "Internal server error: failed to confirm transaction state"}`, http.StatusInternalServerError)
			return 
		}
		
		// 5. RESPONDE AO CLIENTE COM SUCESSO
		cardKeys := make([]string, len(cards))
		for i, c := range cards {
			cardKeys[i] = string(c.Key())
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(PurchaseResponse{Cards: cardKeys})
	}
}
//END OF FILE jokenpo/internal/services/shop/api.go