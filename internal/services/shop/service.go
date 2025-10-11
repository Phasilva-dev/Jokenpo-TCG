package shop

import (
	"jokenpo/internal/game/card"
)

// Tipos de mensagens que o ator aceita
type purchaseRequest struct {
	quantity uint64
	reply    chan purchaseResponse
}

type purchaseResponse struct {
	cards []*card.Card
	err   error
}

type ShopService struct {
	shop      *Shop
	requestCh chan purchaseRequest
}

// Construtor
func NewShopService() *ShopService {
	s := &ShopService{
		shop:      NewShop(),
		requestCh: make(chan purchaseRequest),
	}
	go s.run() // inicia ator
	return s
}

// Loop do ator — processa mensagens de forma sequencial
func (s *ShopService) run() {
	for req := range s.requestCh {
		cards, err := s.shop.purchasePackage(req.quantity)
		req.reply <- purchaseResponse{cards: cards, err: err}
	}
}

// API pública: envia uma requisição para o ator
func (s *ShopService) Purchase(quantity uint64) ([]*card.Card, error) {
	reply := make(chan purchaseResponse)
	s.requestCh <- purchaseRequest{quantity: quantity, reply: reply}
	resp := <-reply
	return resp.cards, resp.err
}