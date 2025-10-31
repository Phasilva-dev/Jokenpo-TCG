//START OF FILE jokenpo/internal/cluster/leader.go
package cluster

import (
	"encoding/json"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	consul "github.com/hashicorp/consul/api"
)

const (
	leaderKeyPrefix = "service/%s/leader"
	stateKeyPrefix  = "service/%s/state"
)

type StatefulService interface {
	GetState() interface{}
	SetState(state []byte) error
	OnBecomeLeader()
	OnBecomeFollower()
}

type LeaderElector struct {
	consulManager *ConsulManager
	nodeID        string
	serviceName   string
	leaderKey     string
	stateKey      string
	isLeader      atomic.Bool
}

func NewLeaderElector(serviceName string, manager *ConsulManager, nodeID string) (*LeaderElector, error) {
	if nodeID == "" {
		return nil, fmt.Errorf("nodeID (hostname) não pode ser vazio para o LeaderElector")
	}

	elector := &LeaderElector{
		consulManager: manager,
		nodeID:        nodeID, // Usa o nodeID recebido como parâmetro
		serviceName:   serviceName,
		leaderKey:     fmt.Sprintf(leaderKeyPrefix, serviceName),
		stateKey:      fmt.Sprintf(stateKeyPrefix, serviceName),
	}
	elector.isLeader.Store(false)
	return elector, nil
}

func (e *LeaderElector) IsLeader() bool {
	return e.isLeader.Load()
}

func (e *LeaderElector) RunForLeadership(service StatefulService) {
	for {
		log.Printf("[%s Elector] Starting new leadership campaign.", e.serviceName)

		client := e.consulManager.GetClient()
		if client == nil {
			log.Printf("[%s Elector] WARN: Consul client not available. Retrying in 10s...", e.serviceName)
			time.Sleep(10 * time.Second)
			continue
		}

		lockLostCh, err := e.acquireLock(client)
		if err != nil {
			log.Printf("[%s Elector] Failed to acquire lock: %v. Retrying in 10s...", e.serviceName, err)
			service.OnBecomeFollower()
			e.isLeader.Store(false)
			time.Sleep(10 * time.Second)
			continue
		}

		log.Printf("**************************************************")
		log.Printf("***** This node (%s) is now the LEADER for service '%s' *****", e.nodeID, e.serviceName)
		log.Println("**************************************************")
		e.isLeader.Store(true)
		e.restoreState(service)
		service.OnBecomeLeader()

		<-lockLostCh

		log.Printf("[%s Elector] Leadership lost. Becoming follower.", e.serviceName)
		service.OnBecomeFollower()
		e.isLeader.Store(false)
	}
}

func (e *LeaderElector) acquireLock(client *consul.Client) (<-chan struct{}, error) {
	// 1) Cria uma session explícita com TTL
	se := &consul.SessionEntry{
		Name:     fmt.Sprintf("%s-leader-session", e.serviceName),
		TTL:      "15s",                          // TTL curto; renovaremos
		Behavior: consul.SessionBehaviorRelease, // libera a chave ao expirar
	}

	sessionID, _, err := client.Session().Create(se, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// 2) Cria o lock usando a session criada
	opts := &consul.LockOptions{
		Key:     e.leaderKey,
		Session: sessionID,
		Value:   []byte(e.nodeID),
	}
	lock, err := client.LockOpts(opts)
	if err != nil {
		// tenta destruir a session criada em caso de erro
		_, _ = client.Session().Destroy(sessionID, nil)
		return nil, fmt.Errorf("failed to create lock: %w", err)
	}

	// 3) Tenta adquirir o lock (bloqueante até adquirir ou erro)
	lockCh, err := lock.Lock(nil)
	if err != nil {
		// cleanup
		_ = lock.Unlock()
		_, _ = client.Session().Destroy(sessionID, nil)
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}

	log.Printf("[%s Elector] 🔒 Lock adquirido com sucesso (session=%s) para chave '%s'", e.serviceName, sessionID, e.leaderKey)

	// 4) Goroutine que renova a sessão periodicamente enquanto o lock estiver ativo.
	//    Se a renovação falhar, fazemos unlock e saímos — isso resultará em lockCh sendo fechado.
	go func() {
		ticker := time.NewTicker(5 * time.Second) // renovar bem antes de 15s
		defer ticker.Stop()

		for range ticker.C {
			// Se lockCh estiver fechado, o lock já foi perdido -> exit
			select {
			case <-lockCh:
				// lock foi perdido/fechado, destruímos a session e saimos
				_, _ = client.Session().Destroy(sessionID, nil)
				return
			default:
			}

			// tenta renovar
			renewed, _, err := client.Session().Renew(sessionID, nil)
			if err != nil {
				log.Printf("[%s Elector] ERRO: falha ao renovar session %s: %v — liberando lock.", e.serviceName, sessionID, err)
				// força o unlock — Lock.Unlock pode retornar erro se já perdido
				_ = lock.Unlock()
				// destruir session (melhor esforço)
				_, _ = client.Session().Destroy(sessionID, nil)
				return
			}

			// sanity check (opcional)
			if renewed == nil {
				log.Printf("[%s Elector] WARN: renew returned nil for session %s — liberando lock.", e.serviceName, sessionID)
				_ = lock.Unlock()
				_, _ = client.Session().Destroy(sessionID, nil)
				return
			}
		}
	}()

	// Retornamos o canal que será fechado quando o lock for perdido.
	return lockCh, nil
}

func (e *LeaderElector) restoreState(service StatefulService) {
	client := e.consulManager.GetClient()
	if client == nil {
		log.Printf("WARN: Cannot restore state for '%s', Consul client is nil.", e.serviceName)
		return
	}
	kvPair, _, err := client.KV().Get(e.stateKey, nil)
	if err != nil {
		log.Printf("WARN: Could not read state for '%s': %v.", e.serviceName, err)
		return
	}
	if kvPair != nil && len(kvPair.Value) > 0 {
		if err := service.SetState(kvPair.Value); err != nil {
			log.Printf("ERROR: Failed to restore state for '%s': %v.", e.serviceName, err)
		} else {
			log.Printf("[Leader] Restored state for '%s'.", e.serviceName)
		}
	}
}

func (e *LeaderElector) PersistState(service StatefulService) error {
	if !e.IsLeader() {
		return fmt.Errorf("only leader can persist state")
	}
	client := e.consulManager.GetClient()
	if client == nil {
		return fmt.Errorf("cannot persist state for '%s', Consul client is nil", e.serviceName)
	}
	state := service.GetState()
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}
	kvPair := &consul.KVPair{Key: e.stateKey, Value: data}
	_, err = client.KV().Put(kvPair, nil)
	if err != nil {
		return fmt.Errorf("failed to write state to Consul: %w", err)
	}
	log.Printf("[Leader] Persisted state for '%s'.", e.serviceName)
	return nil
}
//END OF FILE jokenpo/internal/cluster/leader.go