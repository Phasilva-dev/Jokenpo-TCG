package api

import (
	"encoding/json"
	"jokenpo/internal/services/cluster"
	"jokenpo/internal/services/shop"
	"log"
	"net/http"
)

// DTOs (Data Transfer Objects) para o contrato da API do Shop
type PurchaseRequest struct {
	Quantity uint64 `json:"quantity"`
}

type PurchaseResponse struct {
	Cards []string `json:"cards,omitempty"`
	Error string   `json:"error,omitempty"`
}

// CreateShopHandler cria o handler HTTP para o ShopService.
// Ele garante que apenas o líder processe as requisições e que o estado
// seja persistido antes de confirmar a operação para o cliente.
func CreateShopHandler(shopService *shop.ShopService, elector *cluster.LeaderElector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. VERIFICAÇÃO DE LIDERANÇA
		// Garante que apenas o nó líder possa processar operações de escrita.
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

		// 3. EXECUTA A LÓGICA DE NEGÓCIO (EM MEMÓRIA)
		// A compra é processada e o estado do ator é atualizado em memória.
		cards, err := shopService.Purchase(req.Quantity)
		if err != nil {
			// Se a lógica de negócio falhar (ex: limite de compras), o erro é retornado
			// imediatamente. Nenhuma persistência é necessária.
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(PurchaseResponse{Error: err.Error()})
			return
		}

		// 4. PERSISTE O ESTADO (WRITE-AHEAD)
		// ANTES de responder ao cliente, tentamos salvar o novo estado de forma durável no Consul.
		// Esta é a etapa que garante a consistência.
		if err := elector.PersistState(shopService); err != nil {
			// Cenário crítico: o estado mudou em memória, mas não foi salvo.
			// A transação DEVE ser considerada falha do ponto de vista do cliente.
			// Não podemos enviar as cartas, pois isso criaria uma inconsistência.
			log.Printf("CRITICAL: State changed in memory but failed to persist to Consul: %v. Transaction will NOT be confirmed.", err)

			// Em um sistema mais avançado, aqui seria o local para uma lógica de "rollback"
			// para reverter a mudança em memória.

			http.Error(w, `{"error": "Internal server error: failed to confirm transaction state"}`, http.StatusInternalServerError)
			return // NÃO enviamos a resposta de sucesso.
		}
		
		// 5. RESPONDE AO CLIENTE COM SUCESSO
		// Apenas depois que o estado foi salvo com segurança, nós confirmamos
		// a operação para o cliente e enviamos os dados. Se o servidor travar aqui,
		// o cliente receberá um erro de rede e a transação será considerada falha por ele,
		// o que leva ao cenário do "pacote fantasma" (risco aceitável e mínimo).
		cardKeys := make([]string, len(cards))
		for i, c := range cards {
			cardKeys[i] = string(c.Key())
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(PurchaseResponse{Cards: cardKeys})
	}
}