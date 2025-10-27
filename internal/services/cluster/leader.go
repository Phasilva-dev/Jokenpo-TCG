//START OF FILE jokenpo/internal/cluster/leader.go
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
	client      *consul.Client
	nodeID      string
	serviceName string
	leaderKey   string
	stateKey    string
	isLeader    atomic.Int32
}

// NewLeaderElector cria um novo eleitor, usando um cliente Consul resiliente.
func NewLeaderElector(serviceName string, consulAddrs string) (*LeaderElector, error) {
	// --- MUDANÇA: Usa a nova função helper para criar um cliente resiliente ---
	client, err := NewConsulClient(consulAddrs)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %w", err)
	}

	// O resto da função permanece o mesmo.
	nodeID, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname: %w", err)
	}

	elector := &LeaderElector{
		client:      client,
		nodeID:      nodeID,
		serviceName: serviceName,
		leaderKey:   fmt.Sprintf(leaderKeyPrefix, serviceName),
		stateKey:    fmt.Sprintf(stateKeyPrefix, serviceName),
	}
	elector.isLeader.Store(0)
	return elector, nil
}

// (O resto do arquivo leader.go - IsLeader, RunForLeadership, etc. - não precisa de mudanças)
func (e *LeaderElector) IsLeader() bool {
	return e.isLeader.Load() == 1
}

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
		log.Printf("**************************************************")
		log.Printf("***** This node (%s) is now the LEADER for service '%s' *****", e.nodeID, e.serviceName)
		log.Println("**************************************************")
		e.isLeader.Store(1)
		e.restoreState(service)
		service.OnBecomeLeader()
		<-lockLostCh
		log.Printf("[%s Elector] Leadership lost. Becoming follower.", e.serviceName)
		service.OnBecomeFollower()
		e.isLeader.Store(0)
	}
}

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

func (e *LeaderElector) restoreState(service StatefulService) {
	kvPair, _, err := e.client.KV().Get(e.stateKey, nil)
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
	state := service.GetState()
	data, err := json.Marshal(state)
	if err != nil { return fmt.Errorf("failed to marshal state: %w", err) }
	kvPair := &consul.KVPair{Key: e.stateKey, Value: data}
	_, err = e.client.KV().Put(kvPair, nil)
	if err != nil { return fmt.Errorf("failed to write state to Consul: %w", err) }
	log.Printf("[Leader] Persisted state for '%s'.", e.serviceName)
	return nil
}
//END OF FILE jokenpo/internal/cluster/leader.go