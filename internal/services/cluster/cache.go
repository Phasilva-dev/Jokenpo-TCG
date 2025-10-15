package cluster

import (
	"fmt"
	"time"
)

// serviceCacheEntry armazena o endereço de um serviço e quando essa informação expira.
type serviceCacheEntry struct {
	address    string
	expiration time.Time
}

// discoveryRequest é a "mensagem" que enviamos para o ator do cache.
// Agora inclui as opções de descoberta para especificar o tipo de consulta.
type discoveryRequest struct {
	serviceName string
	opts        DiscoveryOptions // A struct de opções completa
	reply       chan<- string
}

// ServiceCacheActor é o nosso cache de descoberta de serviço, implementado
// com o padrão Ator para garantir a segurança de concorrência.
type ServiceCacheActor struct {
	// Estado privado do ator
	entries    map[string]serviceCacheEntry
	ttl        time.Duration
	consulAddr string

	// Mailbox (caixa de entrada) do ator
	requestCh chan discoveryRequest
}

// NewServiceCacheActor cria, inicializa e inicia o ator do cache.
func NewServiceCacheActor(ttl time.Duration, consulAddr string) *ServiceCacheActor {
	sc := &ServiceCacheActor{
		entries:    make(map[string]serviceCacheEntry),
		ttl:        ttl,
		requestCh:  make(chan discoveryRequest),
		consulAddr: consulAddr,
	}
	go sc.run()
	return sc
}

// run é o loop principal do ator. Ele processa todos os pedidos de descoberta
// de forma sequencial, garantindo a segurança do cache.
func (sc *ServiceCacheActor) run() {
	for req := range sc.requestCh {
		// A chave do cache agora é mais específica para diferenciar os tipos de consulta.
		// Ex: "jokenpo-shop-ModeLeader-", "jokenpo-session-ModeSpecific-node-123"
		cacheKey := fmt.Sprintf("%s-%d-%s", req.serviceName, req.opts.Mode, req.opts.SpecificID)

		// 1. Tenta encontrar uma entrada válida no cache.
		entry, found := sc.entries[cacheKey]
		if found && time.Now().Before(entry.expiration) {
			req.reply <- entry.address
			continue // Cache hit! Pula para a próxima requisição.
		}

		// 2. Cache miss: Chama a nova função 'Discover' unificada.
		address := Discover(req.serviceName, sc.consulAddr, req.opts)

		// 3. Se a descoberta foi bem-sucedida, armazena no cache.
		if address != "" {
			sc.entries[cacheKey] = serviceCacheEntry{
				address:    address,
				expiration: time.Now().Add(sc.ttl),
			}
		}

		// 4. Envia a resposta (seja do cache ou da nova descoberta) de volta.
		req.reply <- address
	}
}

// Discover é a API pública. Aceita o nome do serviço e as opções de descoberta.
func (sc *ServiceCacheActor) Discover(serviceName string, opts DiscoveryOptions) string {
	replyCh := make(chan string)

	// Cria a mensagem para o ator, incluindo as opções.
	req := discoveryRequest{
		serviceName: serviceName,
		opts:        opts,
		reply:       replyCh,
	}

	// Envia o pedido para a fila do ator.
	sc.requestCh <- req

	// Bloqueia e espera pela resposta do ator.
	return <-replyCh
}