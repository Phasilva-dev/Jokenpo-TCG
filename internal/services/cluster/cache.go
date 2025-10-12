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

// ServiceCacheActor é a nossa implementação de cache usando o padrão Ator.
type ServiceCacheActor struct {
	// Estado privado do ator (só a goroutine 'run' pode acessar)
	entries map[string]serviceCacheEntry
	ttl     time.Duration

	// O "mailbox" do ator, onde ele recebe os pedidos.
	requestCh chan discoveryRequest
}

// NewServiceCacheActor cria, inicializa e inicia o ator do cache.
func NewServiceCacheActor(ttl time.Duration) *ServiceCacheActor {
	sc := &ServiceCacheActor{
		entries:   make(map[string]serviceCacheEntry),
		ttl:       ttl,
		requestCh: make(chan discoveryRequest),
	}
	// Inicia a goroutine do ator para que ela comece a ouvir por pedidos.
	go sc.run()
	return sc
}

// run é o loop principal do ator. Ele roda em sua própria goroutine e é o
// único que tem permissão para acessar o mapa 'entries'. Isso elimina
// completamente a necessidade de mutexes.
func (sc *ServiceCacheActor) run() {
	// O ator processa um pedido de cada vez, do início ao fim.
	for req := range sc.requestCh {
		// 1. Verifica se existe uma entrada válida no cache.
		entry, found := sc.entries[req.serviceName]
		if found && time.Now().Before(entry.expiration) {
			// CACHE HIT: Envia o endereço do cache de volta para quem pediu.
			req.reply <- entry.address
			continue // Pula para o próximo pedido na fila.
		}

		// CACHE MISS: Se não encontrou ou expirou, busca um novo endereço.
		address := DiscoverService(req.serviceName) // Fala com o Consul

		// Se a descoberta foi bem-sucedida, atualiza o estado interno (o cache).
		if address != "" {
			sc.entries[req.serviceName] = serviceCacheEntry{
				address:    address,
				expiration: time.Now().Add(sc.ttl),
			}
		}

		// Envia a resposta (o novo endereço ou uma string vazia) de volta.
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