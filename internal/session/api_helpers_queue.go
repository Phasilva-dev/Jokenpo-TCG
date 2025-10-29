// START OF FILE jokenpo/internal/session/api_helpers_queue.go
package session

import (
	"bytes"
	"encoding/json"
	"fmt"
	"jokenpo/internal/services/cluster"
	"log"
	"net/http"
	"os"
)

// ============================================================================
// DTOs para Comunicação com o QueueService
// ============================================================================

// EnqueueMatchRequest é o DTO enviado para entrar na fila de partida.
type EnqueueMatchRequest struct {
	PlayerID    string   `json:"playerId"`
	CallbackURL string   `json:"callbackUrl"`
	Deck        []string `json:"deck"`
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
func (h *GameHandler) enterMatchQueue(session *PlayerSession, deckKeys []string) error {
	opts := cluster.DiscoveryOptions{Mode: cluster.ModeLeader}
	log.Printf("[enterMatchQueue] Tentando descobrir o serviço 'jokenpo-queue' com options: %+v", opts)
	queueServiceAddr := h.serviceCache.Discover("jokenpo-queue", opts)
	log.Printf("[enterMatchQueue] queueServiceAddr retornado: '%s'", queueServiceAddr)
	h.serviceCache.PrintEntries()
	if queueServiceAddr == "" {
		return fmt.Errorf("the matchmaking service is currently unavailable")
	}

	gameEventCallbackURL := fmt.Sprintf("http://%s:%d/game-event", h.advertisedHostname, 8080)

	// --- MUDANÇA: O payload agora inclui o deck ---
	payload := EnqueueMatchRequest{
		PlayerID:    session.ID,
		CallbackURL: gameEventCallbackURL,
		Deck:        deckKeys,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to create request payload: %w", err)
	}

	matchFoundCallbackURL := fmt.Sprintf("http://%s:%d/match-found", h.advertisedHostname, 8080)
	
	// Construímos a URL para o serviço da fila e passamos o callback como um query parameter.
	queueURL := fmt.Sprintf("http://%s/queue/match?callback=%s", queueServiceAddr, matchFoundCallbackURL)

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
	log.Printf("[leaveMatchQueue] Tentando descobrir o serviço 'jokenpo-queue' com options: %+v", opts)
	queueServiceAddr := h.serviceCache.Discover("jokenpo-queue", opts)
	h.serviceCache.PrintEntries()
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
	log.Printf("[enterTradeQueue] Tentando descobrir o serviço 'jokenpo-queue' com options: %+v", opts)
	queueServiceAddr := h.serviceCache.Discover("jokenpo-queue", opts)
	h.serviceCache.PrintEntries()
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
	log.Printf("[leaveTradeQueue] Tentando descobrir o serviço 'jokenpo-queue' com options: %+v", opts)
	queueServiceAddr := h.serviceCache.Discover("jokenpo-queue", opts)
	h.serviceCache.PrintEntries()
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

//END OF FILE jokenpo/internal/session/api_helpers_queue.go