//START OF FILE jokenpo/internal/services/shop/service.go
package shop

import (
	"encoding/json"
	"errors"
	"fmt"
	"jokenpo/internal/game/card"
	"jokenpo/internal/services/blockchain"
	"jokenpo/internal/services/cluster"
	"log"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

type actorMessage interface{ isActorMessage() }

type purchaseRequest struct {
	playerID string
	quantity uint64
	reply    chan purchaseResponse
}
type purchaseResponse struct {
	cards []*card.Card
	err   error
}

func (purchaseRequest) isActorMessage() {}

type healthCheckRequest struct{ reply chan error }

func (healthCheckRequest) isActorMessage() {}

type setStateRequest struct{ newState State }

func (setStateRequest) isActorMessage() {}

type getStateRequest struct{ reply chan State }

func (getStateRequest) isActorMessage() {}

type ShopService struct {
	shop          *Shop
	requestCh     chan actorMessage
	isLeader      atomic.Bool
	blockchain    *blockchain.BlockchainClient
	consulManager *cluster.ConsulManager
}

// NewShopService inicializa o serviço.
// Aguarda o endereço do contrato no Consul (criado pelo Deployer).
func NewShopService(manager *cluster.ConsulManager) *ShopService {
	var bcClient *blockchain.BlockchainClient
	var contractAddr string

	// --- LÓGICA DE ESPERA (POLLING) ---
	log.Println("SHOP: Aguardando endereço do contrato no Consul (via Deployer)...")
	client := manager.GetClient()

	// Tenta por 60 segundos (30 tentativas de 2s)
	for i := 0; i < 30; i++ {
		if client == nil {
			client = manager.GetClient()
		}
		if client != nil {
			pair, _, err := client.KV().Get("jokenpo/config/contract_address", nil)
			if err == nil && pair != nil {
				contractAddr = string(pair.Value)
				break // Endereço encontrado!
			}
		}
		time.Sleep(2 * time.Second)
	}

	if contractAddr != "" {
		var err error
		// MODO CONNECT: Passa o endereço existente para inicializar
		bcClient, _, err = blockchain.InitBlockchain(contractAddr)
		if err != nil {
			log.Printf("SHOP AVISO: Erro ao conectar na blockchain: %v", err)
		} else {
			log.Printf("SHOP: Conectado à Blockchain no contrato: %s", contractAddr)
		}
	} else {
		log.Println("SHOP AVISO: Timeout aguardando contrato. Auditoria desabilitada.")
	}

	s := &ShopService{
		shop:          NewShop(),
		requestCh:     make(chan actorMessage),
		blockchain:    bcClient,
		consulManager: manager,
	}
	s.isLeader.Store(false)
	go s.run()
	return s
}

func (s *ShopService) run() {
	err := card.InitGlobalCatalog()
	if err != nil {
		log.Fatalf("Error catalog: %s", err)
	}
	for msg := range s.requestCh {
		switch req := msg.(type) {
		case purchaseRequest:
			// 1. Tenta comprar localmente (Gera cartas e incrementa contador state.PackageCount)
			cards, err := s.shop.purchasePackage(req.quantity)

			// 2. Se sucesso local, tenta registrar na Blockchain (Síncrono)
			if err == nil && s.blockchain != nil {
				if mintErr := s.mintOnBlockchain(req.playerID, cards); mintErr != nil {
					log.Printf("SHOP CRÍTICO: Blockchain rejeitou transação (%v). Executando Rollback.", mintErr)

					// --- ROLLBACK LÓGICO ---
					// Como purchasePackage() já incrementou o contador, precisamos decrementar
					// para manter a consistência entre Estado Local e Blockchain.
					currentState := s.shop.GetState()
					if currentState.PackageCount >= req.quantity {
						currentState.PackageCount -= req.quantity
						s.shop.SetState(currentState) // Restaura estado anterior
						log.Printf("SHOP ROLLBACK: PackageCount revertido para %d", currentState.PackageCount)
					}

					// Invalida a resposta para o usuário
					cards = nil
					err = fmt.Errorf("falha na auditoria blockchain: %v. Compra cancelada", mintErr)
				}
			} else if err == nil && s.blockchain == nil {
				// Opcional: Se a blockchain estiver fora, você pode decidir se
				// permite a compra (sem auditoria) ou bloqueia.
				// Aqui estamos permitindo, mas logando aviso.
				log.Println("SHOP AVISO: Compra realizada SEM registro na blockchain (serviço offline).")
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

// mintOnBlockchain agora retorna erro para permitir controle de fluxo
func (s *ShopService) mintOnBlockchain(playerID string, cards []*card.Card) error {
	uniqueTokens := make([]string, len(cards))
	for i, c := range cards {
		// Gera um UUID para cada carta
		uniqueTokens[i] = fmt.Sprintf("%s#%s", c.Key(), uuid.NewString())
	}

	// Chama a blockchain e espera a mineração (WaitMined já está no client.go)
	if err := s.blockchain.LogPack(playerID, uniqueTokens); err != nil {
		return err
	}

	log.Printf("SHOP SUCESSO: Pacote registrado para %s", playerID)
	return nil
}

func (s *ShopService) Purchase(playerID string, quantity uint64) ([]*card.Card, error) {
	if !s.isLeader.Load() {
		return nil, errors.New("this node is not the leader")
	}
	reply := make(chan purchaseResponse)
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
		return errors.New("timeout")
	}
}

func (s *ShopService) GetState() interface{} {
	reply := make(chan State)
	s.requestCh <- getStateRequest{reply: reply}
	return <-reply
}

func (s *ShopService) SetState(data []byte) error {
	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return err
	}
	s.requestCh <- setStateRequest{newState: state}
	return nil
}

func (s *ShopService) OnBecomeLeader() {
	log.Println("SHOP: Leader enabled.")
	s.isLeader.Store(true)
}
func (s *ShopService) OnBecomeFollower() {
	log.Println("SHOP: Follower disabled.")
	s.isLeader.Store(false)
}
//END OF FILE jokenpo/internal/services/shop/service.go