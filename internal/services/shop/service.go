//START OF FILE jokenpo/internal/services/shop/service.go
package shop

import (
	"encoding/json"
	"errors"
	"fmt"
	"jokenpo/internal/game/card"
	"jokenpo/internal/services/blockchain"
	"log"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

// actorMessage define o contrato para mensagens enviadas ao ator do ShopService.
type actorMessage interface {
	isActorMessage()
}

// --- Definições de Mensagens ---

// Mensagens para a lógica de negócio principal
type purchaseRequest struct {
	playerID string // Novo campo: Precisamos saber quem está comprando para registrar na blockchain
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
	isLeader   atomic.Bool
	blockchain *blockchain.BlockchainClient // Cliente para registrar na ledger
}

// NewShopService cria a instância do serviço, inicializando-o como seguidor.
func NewShopService() *ShopService {
	// Tenta conectar na blockchain. Se falhar (ex: dev mode sem container), loga mas não quebra.
	bc, err := blockchain.NewBlockchainClient()
	if err != nil {
		log.Printf("SHOP AVISO: Blockchain indisponível: %v. Compras não serão auditadas.", err)
	}

	s := &ShopService{
		shop:       NewShop(),
		requestCh:  make(chan actorMessage),
		blockchain: bc,
	}
	s.isLeader.Store(false) // Começa como seguidor por padrão
	go s.run()
	return s
}

// run é o loop principal do ator, que agora lida com todos os tipos de mensagem.
func (s *ShopService) run() {
	err := card.InitGlobalCatalog()
	if err != nil {
		log.Fatalf("Error ao iniciar o catalogo: %s", err)
	}
	for msg := range s.requestCh {
		switch req := msg.(type) {
		case purchaseRequest:
			// 1. Gera as cartas (Lógica de Negócio)
			cards, err := s.shop.purchasePackage(req.quantity)

			// 2. Registra na Blockchain (Lógica de Auditoria)
			// Só registramos se a compra foi bem sucedida e temos conexão com a blockchain.
			if err == nil && s.blockchain != nil {
				// Executa em goroutine para não bloquear o ator do Shop enquanto espera a mineração
				go s.mintOnBlockchain(req.playerID, cards)
			}

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

// mintOnBlockchain gera UUIDs únicos para as cartas e envia para o contrato.
func (s *ShopService) mintOnBlockchain(playerID string, cards []*card.Card) {
	uniqueTokens := make([]string, len(cards))
	for i, c := range cards {
		// Gera um Token Único: "chave_da_carta#UUID"
		// Ex: "rock:5:red#a1b2-c3d4..."
		// Isso garante que cada carta seja um ativo único na blockchain.
		uniqueTokens[i] = fmt.Sprintf("%s#%s", c.Key(), uuid.NewString())
	}

	if err := s.blockchain.LogPack(playerID, uniqueTokens); err != nil {
		log.Printf("SHOP ERRO: Falha ao registrar mintagem na blockchain: %v", err)
	} else {
		log.Printf("SHOP SUCESSO: Pacote registrado para %s na blockchain. Total: %d cartas.", playerID, len(cards))
	}
}

// --- APIs Públicas ---

// Purchase agora inclui uma verificação interna de liderança para segurança extra.
func (s *ShopService) Purchase(playerID string, quantity uint64) ([]*card.Card, error) {
	if !s.isLeader.Load() {
		return nil, errors.New("this node is not the leader and cannot process purchases")
	}

	reply := make(chan purchaseResponse)
	// Envia o playerID junto com a requisição
	s.requestCh <- purchaseRequest{playerID: playerID, quantity: quantity, reply: reply}
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
//END OF FILE jokenpo/internal/services/shop/service.go