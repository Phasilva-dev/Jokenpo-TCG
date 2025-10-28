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
	addrs       string
	currentAddr string
	client      *consul.Client
	mu          sync.RWMutex
}

// NewConsulManager cria e inicia um novo gerenciador de conexão Consul.
func NewConsulManager(addrs string) (*ConsulManager, error) {
	m := &ConsulManager{
		addrs: addrs,
	}

	// Tenta a conexão inicial.
	if err := m.reconnect(); err != nil {
		return nil, err
	}

	// Inicia uma goroutine para monitorar a saúde da conexão.
	go m.monitor()

	return m, nil
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
	defer m.mu.Unlock()

	log.Println("[ConsulManager] Tentando reconectar ao cluster Consul...")

	nodes := strings.Split(m.addrs, ",")
	for _, node := range nodes {
		node = strings.TrimSpace(node)
		client, err := NewConsulClient(node)
		if err != nil {
			log.Printf("[ConsulManager] Falha ao conectar em %s: %v", node, err)
			continue
		}

		m.client = client
		m.currentAddr = node
		log.Printf("[ConsulManager] Conectado com sucesso ao nó %s", node)
		return nil
	}

	log.Printf("[ConsulManager] Nenhum nó disponível em %s", m.addrs)
	return fmt.Errorf("não foi possível se conectar a nenhum nó Consul")
}

// monitor verifica periodicamente a saúde da conexão e tenta reconectar se necessário.
func (m *ConsulManager) monitor() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		client := m.GetClient()
		if client == nil {
			log.Println("[ConsulManager] Client é nulo, tentando reconectar.")
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