//START OF FILE jokenpo/internal/services/cluster/cache.go
package cluster

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type serviceCacheEntry struct {
	address    string
	expiration time.Time
}

type discoveryRequest struct {
	serviceName string
	opts        DiscoveryOptions
	reply       chan<- string
}

type ServiceCacheActor struct {
	entries       map[string]serviceCacheEntry
	ttl           time.Duration
	consulManager *ConsulManager
	requestCh     chan discoveryRequest
	mu            sync.RWMutex
}

func NewServiceCacheActor(ttl time.Duration, manager *ConsulManager) *ServiceCacheActor {
	sc := &ServiceCacheActor{
		entries:       make(map[string]serviceCacheEntry),
		ttl:           ttl,
		consulManager: manager,
		requestCh:     make(chan discoveryRequest),
	}
	go sc.run()
	return sc
}

func (sc *ServiceCacheActor) run() {
	for req := range sc.requestCh {
		address := sc.internalDiscover(req.serviceName, req.opts)
		req.reply <- address
	}
}

func (sc *ServiceCacheActor) internalDiscover(serviceName string, opts DiscoveryOptions) string {
	cacheKey := fmt.Sprintf("%s-%d-%s", serviceName, opts.Mode, opts.SpecificID)

	// Primeiro tenta o cache
	sc.mu.RLock()
	entry, found := sc.entries[cacheKey]
	sc.mu.RUnlock()

	if found && time.Now().Before(entry.expiration) {
		return entry.address
	}

	// Se não encontrou ou expirou, consulta o Consul
	client := sc.consulManager.GetClient()
	if client == nil {
		log.Printf("[ServiceCache] WARN: Consul client not available for '%s'", serviceName)
		return ""
	}

	address := discoverWithClient(client, serviceName, opts)
	if address != "" {
		sc.mu.Lock()
		sc.entries[cacheKey] = serviceCacheEntry{
			address:    address,
			expiration: time.Now().Add(sc.ttl),
		}
		sc.mu.Unlock()
		log.Printf("[ServiceCache] Updated cache for service '%s': %s", serviceName, address)
	} else {
		log.Printf("[ServiceCache] No healthy address found for service '%s'", serviceName)
	}

	return address
}

// Discover retorna o endereço de um serviço, usando cache se possível
func (sc *ServiceCacheActor) Discover(serviceName string, opts DiscoveryOptions) string {
	replyCh := make(chan string)
	sc.requestCh <- discoveryRequest{
		serviceName: serviceName,
		opts:        opts,
		reply:       replyCh,
	}
	return <-replyCh
}

// Refresh força a atualização do cache para um serviço específico
func (sc *ServiceCacheActor) Refresh(serviceName string, opts DiscoveryOptions) {
	go func() {
		address := sc.internalDiscover(serviceName, opts)
		if address != "" {
			log.Printf("[ServiceCache] Refresh successful for '%s': %s", serviceName, address)
		} else {
			log.Printf("[ServiceCache] Refresh failed for '%s': no address found", serviceName)
		}
	}()
}

func (sc *ServiceCacheActor) PrintEntries() {
    log.Println("[ServiceCache] Entradas atuais no cache:")
    for k, v := range sc.entries {
        log.Printf("  %s -> %s (expira em %v)", k, v.address, v.expiration)
    }
}
//END OF FILE jokenpo/internal/services/cluster/cache.go