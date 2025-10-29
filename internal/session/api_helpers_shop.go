// START OF FILE jokenpo/internal/session/api_helpers_shop.go
package session

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"jokenpo/internal/services/cluster"
	"log"
	"net/http"
)

//DTOs

//SHOP SERVICE
type PurchaseRequest struct {
	Quantity uint64 `json:"quantity"`
}

type PurchaseResponse struct {
	Cards []string `json:"cards"`
	Error string   `json:"error,omitempty"`
}







//SHOP SERVICE
// purchasePacksFromShop é um helper privado do GameHandler que encapsula a lógica
// de comunicação com o ShopService. Retorna as chaves das cartas ou um erro.
// Ele lida com Service Discovery (via Cache) e chamadas HTTP (via Client compartilhado).
func (h *GameHandler) purchasePacksFromShop(quantity uint64) ([]string, error) {
	// --- MUDANÇA ---
	// 1. Especifica que queremos encontrar o LÍDER do cluster do shop.
	opts := cluster.DiscoveryOptions{Mode: cluster.ModeLeader}
	log.Printf("[purchasePacksFromShop] Tentando descobrir o serviço 'jokenpo-queue' com options: %+v", opts)
	shopAddr := h.serviceCache.Discover("jokenpo-shop", opts)
	h.serviceCache.PrintEntries()

	if shopAddr == "" {
		// A mensagem de erro agora é mais precisa e acionável.
		return nil, fmt.Errorf("não foi possível encontrar o líder do shop service no momento")
	}

	// 2. Prepara a chamada HTTP
	shopURL := fmt.Sprintf("http://%s/Purchase", shopAddr)

	// --- MUDANÇA (Melhor Prática) ---
	// Usa a struct DTO para criar o payload, garantindo consistência com a API.
	reqPayload := PurchaseRequest{Quantity: quantity}
	reqBody, err := json.Marshal(reqPayload)
	if err != nil {
		// Embora improvável, é bom tratar este erro.
		return nil, fmt.Errorf("falha ao serializar o payload da requisição: %w", err)
	}

	httpReq, err := http.NewRequest("POST", shopURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// 3. Executa a chamada (lógica inalterada)
	resp, err := h.httpClient.Do(httpReq)
	if err != nil {
		// A mensagem de erro agora pode ser mais específica.
		return nil, fmt.Errorf("failed to contact shop service leader at %s: %w", shopAddr, err)
	}
	defer resp.Body.Close()

	// 4. Processa a resposta (lógica inalterada)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var shopResp PurchaseResponse
	if err := json.Unmarshal(body, &shopResp); err != nil {
		return nil, fmt.Errorf("failed to parse response from shop service: %w", err)
	}
	if shopResp.Error != "" {
		return nil, fmt.Errorf("shop service error: %s", shopResp.Error)
	}

	return shopResp.Cards, nil
}

//END OF FILE jokenpo/internal/session/api_helpers_shop.go