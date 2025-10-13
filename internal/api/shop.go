package api

import (
	"encoding/json"
	"jokenpo/internal/services/cluster" // <-- 1. IMPORTA o pacote genérico de cluster
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

// --- MUDANÇA ---
// A função agora aceita um 'elector' para gerenciar a lógica de liderança.
func CreateShopHandler(shopService *shop.ShopService, elector *cluster.LeaderElector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// --- MUDANÇA ---
		// 1. VERIFICAÇÃO DE LIDERANÇA
		// Antes de qualquer outra coisa, verifica se este nó é o líder.
		// Esta é a "porta de entrada" para a lógica de escrita.
		if !elector.IsLeader() {
			// Retorna um erro HTTP 503 (Service Unavailable). Isso sinaliza ao cliente
			// que o serviço está temporariamente indisponível (neste caso, porque é um seguidor).
			// Balanceadores de carga podem usar este status para tentar outro nó.
			http.Error(w, `{"error": "This node is not the leader and cannot process write operations"}`, http.StatusServiceUnavailable)
			return
		}

		// 2. Decodifica o JSON vindo do Broker
		var req PurchaseRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "Invalid payload"}`, http.StatusBadRequest)
			return
		}

		// 3. Chama a lógica de negócio real
		// O método Purchase() do shopService agora também tem uma verificação interna,
		// o que nos dá segurança em camadas.
		cards, err := shopService.Purchase(req.Quantity)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(PurchaseResponse{Error: err.Error()})
			return
		}

		// --- MUDANÇA ---
		// 4. PERSISTÊNCIA DE ESTADO
		// Após a compra ser bem-sucedida, o líder DEVE persistir o novo estado.
		// Isso é CRÍTICO para a consistência em caso de falha.
		if err := elector.PersistState(shopService); err != nil {
			// Este é um erro grave. A compra aconteceu, mas o estado não foi salvo,
			// criando uma inconsistência se o líder falhar agora.
			// Em produção, isso dispararia um alerta de alta prioridade.
			log.Printf("CRITICAL: Purchase successful but failed to persist state to Consul: %v", err)
			// Apesar do erro de persistência, a operação do ponto de vista do
			// cliente foi um sucesso, então ainda retornamos 200 OK.
		}

		// 5. Converte o resultado para o DTO de resposta
		cardKeys := make([]string, len(cards))
		for i, c := range cards {
			cardKeys[i] = string(c.Key())
		}

		// 6. Envia a resposta de sucesso de volta para o Broker
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(PurchaseResponse{Cards: cardKeys})
	}
}