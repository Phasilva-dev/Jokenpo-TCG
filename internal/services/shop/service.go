package shop

import (
	"errors"
	"jokenpo/internal/game/card"
	"time"
)

// actorMessage define o contrato para mensagens enviadas ao ator do ShopService.
type actorMessage interface {
	isActorMessage()
}

// --- Definições de Mensagens ---

type purchaseRequest struct {
	quantity uint64
	reply    chan purchaseResponse
}
type purchaseResponse struct {
	cards []*card.Card
	err   error
}
func (purchaseRequest) isActorMessage() {} // Satisfaz a interface

type healthCheckRequest struct {
	reply chan error
}
func (healthCheckRequest) isActorMessage() {} // Satisfaz a interface

// --- Implementação do Ator ---

type ShopService struct {
	shop *Shop
	// O canal agora é fortemente tipado para a nossa interface de mensagem.
	requestCh chan actorMessage
}

func NewShopService() *ShopService {
	s := &ShopService{
		shop:      NewShop(),
		requestCh: make(chan actorMessage), // Canal fortemente tipado
	}
	go s.run()
	return s
}

// O loop do ator permanece o mesmo, mas agora ele tem a garantia de que
// só receberá tipos que satisfaçam a interface actorMessage.
func (s *ShopService) run() {
	for msg := range s.requestCh {
		switch req := msg.(type) {
		case purchaseRequest:
			cards, err := s.shop.purchasePackage(req.quantity)
			req.reply <- purchaseResponse{cards: cards, err: err}

		case healthCheckRequest:
			req.reply <- nil
		}
	}
}

// --- APIs Públicas ---

func (s *ShopService) Purchase(quantity uint64) ([]*card.Card, error) {
	reply := make(chan purchaseResponse)
	// Criamos a mensagem e a enviamos. O compilador garante que é um tipo válido.
	s.requestCh <- purchaseRequest{quantity: quantity, reply: reply}
	resp := <-reply
	return resp.cards, resp.err
}

func (s *ShopService) CheckHealth() error {
	reply := make(chan error)
	// Criamos a mensagem e a enviamos. O compilador também garante que é válido.
	s.requestCh <- healthCheckRequest{reply: reply}

	select {
	case err := <-reply:
		return err
	case <-time.After(1 * time.Second):
		return errors.New("health check timed out: actor goroutine is unresponsive")
	}
}