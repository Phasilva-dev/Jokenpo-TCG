package cluster

import (
	"time"
)

type serviceCacheEntry struct {
	address    string
	expiration time.Time
}

// discoveryRequest é a "mensagem" que enviamos para o ator do cache.
// Contém o que queremos (serviceName) e como obter a resposta (reply channel).
type discoveryRequest struct {
	serviceName string
	reply       chan<- string // Canal de mão única para enviar a resposta
}

// ServiceCacheActor agora armazena o endereço do Consul.
type ServiceCacheActor struct {
	// Estado privado
	entries    map[string]serviceCacheEntry
	ttl        time.Duration
	consulAddr string // MODIFICADO: Adicionado campo para o endereço do Consul

	// Mailbox
	requestCh chan discoveryRequest
}

// MODIFICADO: NewServiceCacheActor agora recebe o endereço do Consul.
func NewServiceCacheActor(ttl time.Duration, consulAddr string) *ServiceCacheActor {
	sc := &ServiceCacheActor{
		entries:    make(map[string]serviceCacheEntry),
		ttl:        ttl,
		requestCh:  make(chan discoveryRequest),
		consulAddr: consulAddr, // Armazena o endereço
	}
	go sc.run()
	return sc
}

// run agora usa o endereço do Consul armazenado.
func (sc *ServiceCacheActor) run() {
	for req := range sc.requestCh {
		entry, found := sc.entries[req.serviceName]
		if found && time.Now().Before(entry.expiration) {
			req.reply <- entry.address
			continue
		}

		// MODIFICADO: Passa o endereço do Consul para a função de descoberta.
		address := DiscoverService(req.serviceName, sc.consulAddr)

		if address != "" {
			sc.entries[req.serviceName] = serviceCacheEntry{
				address:    address,
				expiration: time.Now().Add(sc.ttl),
			}
		}
		req.reply <- address
	}
}

// Discover é a API pública. É como o resto do programa interage com o ator.
func (sc *ServiceCacheActor) Discover(serviceName string) string {
	// 1. Cria um canal de resposta específico para este pedido.
	replyCh := make(chan string)

	// 2. Cria a mensagem de requisição.
	req := discoveryRequest{
		serviceName: serviceName,
		reply:       replyCh,
	}

	// 3. Envia o pedido para a "caixa de entrada" do ator.
	sc.requestCh <- req

	// 4. Bloqueia e espera até que o ator envie uma resposta de volta.
	address := <-replyCh

	// 5. Retorna a resposta recebida.
	return address
}