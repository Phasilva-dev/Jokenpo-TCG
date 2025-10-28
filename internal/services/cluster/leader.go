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
	consulManager *ConsulManager
	nodeID        string
	serviceName   string
	leaderKey     string
	stateKey      string
	isLeader      atomic.Bool
}

func NewLeaderElector(serviceName string, manager *ConsulManager) (*LeaderElector, error) {
	hostname := os.Getenv("SERVICE_ADVERTISED_HOSTNAME")
	if hostname == "" {
		return nil, fmt.Errorf("SERVICE_ADVERTISED_HOSTNAME environment variable is not set")
	}

	elector := &LeaderElector{
		consulManager: manager,
		nodeID:        hostname,
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
	opts := &consul.LockOptions{
		Key:        e.leaderKey,
		Value:      []byte(e.nodeID),
		SessionTTL: "15s",
	}
	lock, err := client.LockOpts(opts)
	if err != nil {
		return nil, err
	}
	return lock.Lock(nil)
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