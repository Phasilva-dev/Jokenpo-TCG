//START OF FILE jokenpo/internal/services/cluster/cache.go
package cluster

import (
	"fmt"
	"log"
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
		cacheKey := fmt.Sprintf("%s-%d-%s", req.serviceName, req.opts.Mode, req.opts.SpecificID)

		entry, found := sc.entries[cacheKey]
		if found && time.Now().Before(entry.expiration) {
			req.reply <- entry.address
			continue
		}

		client := sc.consulManager.GetClient()
		if client == nil {
			log.Printf("[ServiceCache] WARN: Consul client not available. Service discovery for '%s' will fail.", req.serviceName)
			req.reply <- ""
			continue
		}

		address := discoverWithClient(client, req.serviceName, req.opts)

		if address != "" {
			sc.entries[cacheKey] = serviceCacheEntry{
				address:    address,
				expiration: time.Now().Add(sc.ttl),
			}
		}
		req.reply <- address
	}
}

func (sc *ServiceCacheActor) Discover(serviceName string, opts DiscoveryOptions) string {
	replyCh := make(chan string)
	req := discoveryRequest{
		serviceName: serviceName,
		opts:        opts,
		reply:       replyCh,
	}
	sc.requestCh <- req
	return <-replyCh
}
//END OF FILE jokenpo/internal/services/cluster/cache.go