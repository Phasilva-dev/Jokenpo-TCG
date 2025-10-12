package api
import (
	"encoding/json"
	"jokenpo/internal/services/shop"
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

// CreateShopHandler é o SERVIDOR. Ele recebe a instância do serviço (o ator)
// e retorna um http.HandlerFunc que processa os requests que chegam do Broker.
func CreateShopHandler(shopService *shop.ShopService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Decodifica o JSON vindo do Broker
		var req PurchaseRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "Invalid payload"}`, http.StatusBadRequest)
			return
		}

		// 2. Chama a lógica de negócio real, de forma segura, via ator
		cards, err := shopService.Purchase(req.Quantity)
		if err != nil {
			// Se o serviço retornou um erro, envia-o na resposta
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(PurchaseResponse{Error: err.Error()})
			return
		}

		// 3. Converte o resultado []*card.Card para o DTO de resposta []string
		cardKeys := make([]string, len(cards))
		for i, c := range cards {
			cardKeys[i] = string(c.Key())
		}

		// 4. Envia a resposta de sucesso de volta para o Broker
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(PurchaseResponse{Cards: cardKeys})
	}
}