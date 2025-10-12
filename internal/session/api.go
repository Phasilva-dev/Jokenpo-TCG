package session

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
	// 1. Descobre o ShopService via Cache
	shopAddr := h.serviceCache.Discover("jokenpo-shop")
	if shopAddr == "" {
		return nil, fmt.Errorf("shop service is currently unavailable")
	}

	// 2. Prepara a chamada HTTP
	shopURL := fmt.Sprintf("http://%s/Purchase", shopAddr)
	reqPayload := map[string]uint64{"quantity": quantity}
	reqBody, _ := json.Marshal(reqPayload) // Erro de Marshal é improvável aqui

	httpReq, err := http.NewRequest("POST", shopURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// 3. Executa a chamada
	resp, err := h.httpClient.Do(httpReq)
	if err != nil {
		// Captura timeout, falhas de rede, etc.
		return nil, fmt.Errorf("failed to contact shop service: %w", err)
	}
	defer resp.Body.Close()

	// 4. Processa a resposta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var shopResp PurchaseResponse 
	if err := json.Unmarshal(body, &shopResp); err != nil {
		return nil, fmt.Errorf("failed to parse response from shop service: %w", err)
	}
	if shopResp.Error != "" {
		// O ShopService retornou um erro de negócio ou interno
		return nil, fmt.Errorf("shop service error: %s", shopResp.Error)
	}

	return shopResp.Cards, nil
}