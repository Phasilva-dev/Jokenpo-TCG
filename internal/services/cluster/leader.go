package cluster

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync/atomic"
	"time"

	consul "github.com/hashicorp/consul/api"
)

const (
	leaderKeyPrefix = "service/%s/leader" // %s será o nome do serviço
	stateKeyPrefix  = "service/%s/state"  // %s será o nome do serviço
)

// StatefulService é a interface que qualquer serviço deve implementar
// para ser gerenciado pelo LeaderElector.
type StatefulService interface {
	// GetState é chamado para obter o estado atual do serviço para persistência.
	// O retorno deve ser serializável em JSON.
	GetState() interface{}

	// SetState é chamado quando o nó se torna líder para restaurar o estado
	// lido do Consul. O serviço é responsável por fazer o type assertion
	// do 'state' para sua struct de estado concreta.
	SetState(state []byte) error

	// OnBecomeLeader é um callback chamado quando o nó ganha a liderança.
	OnBecomeLeader()

	// OnBecomeFollower é um callback chamado quando o nó perde a liderança.
	OnBecomeFollower()
}

// LeaderElector gerencia o processo de eleição para um serviço genérico.
type LeaderElector struct {
	client     *consul.Client
	nodeID     string
	serviceName string
	leaderKey   string
	stateKey    string

	isLeader atomic.Int32
}

// NewLeaderElector cria um novo eleitor para um serviço específico.
func NewLeaderElector(serviceName string) (*LeaderElector, error) {
	config := consul.DefaultConfig()
	config.Address = os.Getenv("CONSUL_HTTP_ADDR")
	if config.Address == "" {
		config.Address = "consul:8500"
	}

	client, err := consul.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %w", err)
	}

	nodeID, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname: %w", err)
	}

	elector := &LeaderElector{
		client:      client,
		nodeID:      fmt.Sprintf("%s-%d", nodeID, os.Getpid()),
		serviceName: serviceName,
		leaderKey:   fmt.Sprintf(leaderKeyPrefix, serviceName),
		stateKey:    fmt.Sprintf(stateKeyPrefix, serviceName),
	}
	elector.isLeader.Store(0)
	return elector, nil
}

// IsLeader retorna true se a instância atual for o líder.
func (e *LeaderElector) IsLeader() bool {
	return e.isLeader.Load() == 1
}

// RunForLeadership inicia o loop de eleição para um serviço que implementa a interface.
func (e *LeaderElector) RunForLeadership(service StatefulService) {
	for {
		log.Printf("[%s Elector] Starting new leadership campaign.", e.serviceName)

		lockLostCh, err := e.acquireLock()
		if err != nil {
			log.Printf("[%s Elector] Failed to acquire lock: %v. Retrying in 10s...", e.serviceName, err)
			service.OnBecomeFollower()
			e.isLeader.Store(0)
			time.Sleep(10 * time.Second)
			continue
		}

		// SUCESSO! SOMOS O LÍDER!
		log.Printf("**************************************************")
		log.Printf("***** This node (%s) is now the LEADER for service '%s' *****", e.nodeID, e.serviceName)
		log.Println("**************************************************")
		e.isLeader.Store(1)

		// Restaura o estado antes de anunciar que é o líder.
		e.restoreState(service)
		
		// Chama o callback para o serviço saber que é o líder.
		service.OnBecomeLeader()

		// Bloqueia até a liderança ser perdida.
		<-lockLostCh
		log.Printf("[%s Elector] Leadership lost. Becoming follower.", e.serviceName)
		service.OnBecomeFollower()
		e.isLeader.Store(0)
	}
}

// acquireLock tenta obter o lock no Consul.
func (e *LeaderElector) acquireLock() (<-chan struct{}, error) {
	opts := &consul.LockOptions{
		Key:        e.leaderKey,
		Value:      []byte(e.nodeID),
		SessionTTL: "15s",
	}
	lock, err := e.client.LockOpts(opts)
	if err != nil {
		return nil, err
	}
	return lock.Lock(nil)
}

// restoreState carrega o último estado do Consul e o aplica ao serviço.
func (e *LeaderElector) restoreState(service StatefulService) {
	kvPair, _, err := e.client.KV().Get(e.stateKey, nil)
	if err != nil {
		log.Printf("WARN: Could not read previous state for service '%s' from Consul KV: %v. Starting fresh.", e.serviceName, err)
		return
	}

	if kvPair != nil && len(kvPair.Value) > 0 {
		if err := service.SetState(kvPair.Value); err != nil {
			log.Printf("ERROR: Failed to restore state for service '%s': %v.", e.serviceName, err)
		} else {
			log.Printf("[Leader] Successfully restored state for service '%s' from Consul KV.", e.serviceName)
		}
	}
}

// PersistState é uma função que o LÍDER deve chamar para salvar o estado no Consul.
func (e *LeaderElector) PersistState(service StatefulService) error {
	if !e.IsLeader() {
		return fmt.Errorf("only the leader can persist state")
	}

	state := service.GetState()
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	kvPair := &consul.KVPair{
		Key:   e.stateKey,
		Value: data,
	}

	_, err = e.client.KV().Put(kvPair, nil)
	if err != nil {
		return fmt.Errorf("failed to write state to Consul KV: %w", err)
	}
	
	log.Printf("[Leader] Persisted state for service '%s' to Consul KV.", e.serviceName)
	return nil
}