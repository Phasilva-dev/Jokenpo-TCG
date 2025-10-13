package shop

import (
	"encoding/json"
	"errors"
	"jokenpo/internal/game/card"
	"log"
	"sync/atomic" // Importa o pacote para operações atômicas
	"time"
)

// actorMessage define o contrato para mensagens enviadas ao ator do ShopService.
type actorMessage interface {
	isActorMessage()
}

// --- Definições de Mensagens ---

// Mensagens para a lógica de negócio principal
type purchaseRequest struct {
	quantity uint64
	reply    chan purchaseResponse
}
type purchaseResponse struct {
	cards []*card.Card
	err   error
}

func (purchaseRequest) isActorMessage() {} // Satisfaz a interface

// Mensagem para health checks
type healthCheckRequest struct {
	reply chan error
}

func (healthCheckRequest) isActorMessage() {} // Satisfaz a interface

// Mensagens para gerenciar o estado para a eleição de líder
type setStateRequest struct {
	newState State
}

func (setStateRequest) isActorMessage() {}

type getStateRequest struct {
	reply chan State
}

func (getStateRequest) isActorMessage() {}

// --- Implementação do Ator ---

type ShopService struct {
	shop *Shop
	// O canal para todas as mensagens do ator
	requestCh chan actorMessage
	// isLeader é um booleano atômico que nos permite verificar de forma
	// segura se esta instância é o líder. 1 = true, 0 = false.
	isLeader atomic.Bool
}

// NewShopService cria a instância do serviço, inicializando-o como seguidor.
func NewShopService() *ShopService {
	s := &ShopService{
		shop:      NewShop(),
		requestCh: make(chan actorMessage),
	}
	s.isLeader.Store(false) // Começa como seguidor por padrão
	go s.run()
	return s
}

// run é o loop principal do ator, que agora lida com todos os tipos de mensagem.
func (s *ShopService) run() {
	for msg := range s.requestCh {
		switch req := msg.(type) {
		case purchaseRequest:
			cards, err := s.shop.purchasePackage(req.quantity)
			req.reply <- purchaseResponse{cards: cards, err: err}

		case healthCheckRequest:
			req.reply <- nil

		case setStateRequest:
			s.shop.SetState(req.newState)

		case getStateRequest:
			req.reply <- s.shop.GetState()
		}
	}
}

// --- APIs Públicas ---

// Purchase agora inclui uma verificação interna de liderança para segurança extra.
func (s *ShopService) Purchase(quantity uint64) ([]*card.Card, error) {
	if !s.isLeader.Load() {
		return nil, errors.New("this node is not the leader and cannot process purchases")
	}

	reply := make(chan purchaseResponse)
	s.requestCh <- purchaseRequest{quantity: quantity, reply: reply}
	resp := <-reply
	return resp.cards, resp.err
}

func (s *ShopService) CheckHealth() error {
	reply := make(chan error)
	s.requestCh <- healthCheckRequest{reply: reply}
	select {
	case err := <-reply:
		return err
	case <-time.After(1 * time.Second):
		return errors.New("health check timed out: actor goroutine is unresponsive")
	}
}

// --- APIs para implementar a interface cluster.StatefulService ---

// GetState obtém o estado atual do ator de forma segura e síncrona.
func (s *ShopService) GetState() interface{} {
	reply := make(chan State)
	s.requestCh <- getStateRequest{reply: reply}
	return <-reply
}

// SetState define o estado do ator de forma segura e síncrona.
func (s *ShopService) SetState(data []byte) error {
	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return err
	}
	s.requestCh <- setStateRequest{newState: state}
	return nil
}

// OnBecomeLeader é o callback chamado pelo eleitor. Ele habilita a lógica de compras.
func (s *ShopService) OnBecomeLeader() {
	log.Println("[ShopService] This instance is now acting as the leader. ENABLING purchase logic.")
	s.isLeader.Store(true)
}

// OnBecomeFollower é o callback chamado pelo eleitor. Ele desabilita a lógica de compras.
func (s *ShopService) OnBecomeFollower() {
	log.Println("[ShopService] This instance is now acting as a follower. DISABLING purchase logic.")
	s.isLeader.Store(false)
}