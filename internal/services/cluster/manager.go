// START OF FILE jokenpo/internal/cluster/manager.go
package cluster

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	consul "github.com/hashicorp/consul/api"
)

// ConsulManager gerencia a conexão com o cluster Consul, garantindo resiliência.
type ConsulManager struct {
	addrs              string
	currentAddr        string
	client             *consul.Client
	mu                 sync.RWMutex
	onReconnectCallbacks []func() // <-- MUDANÇA: Slice para guardar callbacks
}

// NewConsulManager cria e inicia um novo gerenciador de conexão Consul.
func NewConsulManager(addrs string) (*ConsulManager, error) {
	m := &ConsulManager{
		addrs: addrs,
		onReconnectCallbacks: make([]func(), 0), // <-- MUDANÇA: Inicializa o slice
	}

	if err := m.reconnect(); err != nil {
		return nil, err
	}
	go m.monitor()
	return m, nil
}

// OnReconnect registra uma função a ser chamada toda vez que uma reconexão é bem-sucedida.
func (m *ConsulManager) OnReconnect(callback func()) { // <-- MUDANÇA: Nova função
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onReconnectCallbacks = append(m.onReconnectCallbacks, callback)
}

// GetClient retorna um cliente Consul funcional.
func (m *ConsulManager) GetClient() *consul.Client {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.client
}

// reconnect tenta estabelecer uma nova conexão com um nó saudável do Consul.
func (m *ConsulManager) reconnect() error {
	m.mu.Lock()
	defer m.mu.Unlock() // <-- MUDANÇA: Defer agora pode ficar no topo

	log.Println("[ConsulManager] Tentando reconectar ao cluster Consul...")
	m.client = nil

	nodes := strings.Split(m.addrs, ",")
	for _, node := range nodes {
		nodeAddr := strings.TrimSpace(node)
		
		cfg := consul.DefaultConfig()
		cfg.Address = nodeAddr
		client, err := consul.NewClient(cfg)
		if err != nil {
			continue // Tenta o próximo
		}

		// Verifica se o nó pode se comunicar com um líder
		if _, err := client.Status().Leader(); err != nil {
			log.Printf("[ConsulManager] Nó %s indisponível (sem líder): %v", nodeAddr, err)
			continue
		}

		// Sucesso!
		m.client = client
		m.currentAddr = nodeAddr
		log.Printf("[ConsulManager] ✅ Conectado com sucesso ao nó %s", m.currentAddr)
		
		// --- A MÁGICA ---
		// Dispara todos os callbacks de reconexão registrados
		for _, cb := range m.onReconnectCallbacks {
			go cb() // Executa em goroutine para não bloquear o manager
		}

		return nil
	}

	return fmt.Errorf("não foi possível se conectar a nenhum nó Consul em %s", m.addrs)
}

// monitor verifica periodicamente a saúde da conexão e tenta reconectar se necessário.
func (m *ConsulManager) monitor() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		client := m.GetClient()
		if client == nil {
			log.Println("[ConsulManager] Cliente é nulo, tentando reconectar.")
			m.reconnect()
			continue
		}

		if _, err := client.Status().Leader(); err != nil {
			log.Printf("[ConsulManager] WARN: Health check falhou no nó %s: %v. Tentando outros nós.",
				m.currentAddr, err)
			m.reconnect()
		}
	}
}
//END OF FILE jokenpo/internal/cluster/manager.go