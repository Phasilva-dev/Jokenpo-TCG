//START OF FILE jokenpo/internal/session/api_helpers.go
package session

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"jokenpo/internal/services/cluster"
	"net/http"
	"os"
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
	shopAddr := h.serviceCache.Discover("jokenpo-shop", opts)

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



// ============================================================================
// DTOs para Comunicação com o QueueService
// ============================================================================

// EnqueueMatchRequest é o DTO enviado para entrar na fila de partida.
type EnqueueMatchRequest struct {
	PlayerID    string `json:"playerId"`
	CallbackURL string `json:"callbackUrl"`
}

// EnqueueTradeRequest é o DTO enviado para entrar na fila de troca.
type EnqueueTradeRequest struct {
	PlayerID    string `json:"playerId"`
	CallbackURL string `json:"callbackUrl"`
	OfferCard   string `json:"offerCard"`
}

// DequeueRequest é o DTO genérico enviado para sair de qualquer fila.
type DequeueRequest struct {
	PlayerID string `json:"playerId"`
}

// ============================================================================
// Helpers de API para o GameHandler
// ============================================================================

// --- Helpers da Fila de Partida ---

// enterMatchQueue encapsula a chamada HTTP para entrar na fila de partida.
func (h *GameHandler) enterMatchQueue(session *PlayerSession) error {
	// 1. Descobre o LÍDER do serviço de fila.
	opts := cluster.DiscoveryOptions{Mode: cluster.ModeLeader}
	queueServiceAddr := h.serviceCache.Discover("jokenpo-queue", opts)
	if queueServiceAddr == "" {
		return fmt.Errorf("the matchmaking service is currently unavailable")
	}

	// 2. Prepara o payload com o ID da sessão e a URL de callback.
	hostname, _ := os.Hostname()
	callbackURL := fmt.Sprintf("http://%s:%d/match-found", hostname, 8080) // Assume a porta 8080

	payload := EnqueueMatchRequest{
		PlayerID:    session.ID,
		CallbackURL: callbackURL,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to create request payload: %w", err)
	}

	// 3. Faz a chamada POST para o QueueService.
	queueURL := fmt.Sprintf("http://%s/queue/match", queueServiceAddr)
	resp, err := h.httpClient.Post(queueURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to contact matchmaking service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("matchmaking service returned an error status: %s", resp.Status)
	}

	return nil
}

// leaveMatchQueue encapsula a chamada HTTP para sair da fila de partida.
func (h *GameHandler) leaveMatchQueue(session *PlayerSession) error {
	opts := cluster.DiscoveryOptions{Mode: cluster.ModeLeader}
	queueServiceAddr := h.serviceCache.Discover("jokenpo-queue", opts)
	if queueServiceAddr == "" {
		return fmt.Errorf("the matchmaking service is currently unavailable")
	}

	payload := DequeueRequest{PlayerID: session.ID}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to create request payload: %w", err)
	}

	queueURL := fmt.Sprintf("http://%s/queue/match", queueServiceAddr)
	req, err := http.NewRequest(http.MethodDelete, queueURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create DELETE request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to contact matchmaking service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("matchmaking service returned an error status: %s", resp.Status)
	}

	return nil
}

// --- Helpers da Fila de Troca ---

// enterTradeQueue encapsula a chamada HTTP para entrar na fila de troca.
func (h *GameHandler) enterTradeQueue(session *PlayerSession, offerCard string) error {
	opts := cluster.DiscoveryOptions{Mode: cluster.ModeLeader}
	queueServiceAddr := h.serviceCache.Discover("jokenpo-queue", opts)
	if queueServiceAddr == "" {
		return fmt.Errorf("the trade service is currently unavailable")
	}

	hostname, _ := os.Hostname()
	callbackURL := fmt.Sprintf("http://%s:%d/trade-found", hostname, 8080)

	payload := EnqueueTradeRequest{
		PlayerID:    session.ID,
		CallbackURL: callbackURL,
		OfferCard:   offerCard,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to create request payload: %w", err)
	}

	queueURL := fmt.Sprintf("http://%s/queue/trade", queueServiceAddr)
	resp, err := h.httpClient.Post(queueURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to contact trade service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("trade service returned an error status: %s", resp.Status)
	}

	return nil
}

// leaveTradeQueue encapsula a chamada HTTP para sair da fila de troca.
func (h *GameHandler) leaveTradeQueue(session *PlayerSession) error {
	opts := cluster.DiscoveryOptions{Mode: cluster.ModeLeader}
	queueServiceAddr := h.serviceCache.Discover("jokenpo-queue", opts)
	if queueServiceAddr == "" {
		return fmt.Errorf("the trade service is currently unavailable")
	}

	payload := DequeueRequest{PlayerID: session.ID}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to create request payload: %w", err)
	}

	queueURL := fmt.Sprintf("http://%s/queue/trade", queueServiceAddr)
	req, err := http.NewRequest(http.MethodDelete, queueURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create DELETE request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to contact trade service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("trade service returned an error status: %s", resp.Status)
	}

	return nil
}