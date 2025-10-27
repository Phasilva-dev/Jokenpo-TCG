//START OF FILE jokenpo/cmd/server/queue/main.go
package main

import (
	"fmt"
	"jokenpo/internal/services/cluster"
	"jokenpo/internal/services/queue"
	"log"
	"net/http"
	"os"
	"strconv"
)

// ============================================================================
// Constantes de Configuração Padrão
// ============================================================================
const (
	defaultServiceName = "jokenpo-queue"
	defaultServicePort = 8082
	defaultHealthPort  = 8082
	// --- MUDANÇA: O padrão agora é uma lista ---
	defaultConsulAddr = "consul-1:8500,consul-2:8500,consul-3:8500"
)

// ============================================================================
// Estrutura de Configuração
// ============================================================================
type Config struct {
	ServiceName string
	ServicePort int
	HealthPort  int
	ConsulAddrs string // Renomeado para 'Addrs' para indicar que é uma lista
}

// loadConfig carrega a configuração a partir de variáveis de ambiente.
func loadConfig() (*Config, error) {
	serviceName := os.Getenv("QUEUE_SERVICE_NAME")
	if serviceName == "" {
		serviceName = defaultServiceName
	}

	// Lê a lista de endereços do Consul.
	consulAddrs := os.Getenv("CONSUL_HTTP_ADDR")
	if consulAddrs == "" {
		consulAddrs = defaultConsulAddr
	}

	servicePortStr := os.Getenv("QUEUE_SERVICE_PORT")
	if servicePortStr == "" {
		servicePortStr = fmt.Sprintf("%d", defaultServicePort)
	}
	servicePort, err := strconv.Atoi(servicePortStr)
	if err != nil {
		return nil, fmt.Errorf("formato de QUEUE_SERVICE_PORT inválido: %w", err)
	}

	healthPortStr := os.Getenv("HEALTH_CHECK_PORT")
	if healthPortStr == "" {
		healthPortStr = fmt.Sprintf("%d", defaultHealthPort)
	}
	healthPort, err := strconv.Atoi(healthPortStr)
	if err != nil {
		return nil, fmt.Errorf("formato de HEALTH_CHECK_PORT inválido: %w", err)
	}

	return &Config{
		ServiceName: serviceName,
		ServicePort: servicePort,
		HealthPort:  healthPort,
		ConsulAddrs: consulAddrs, // Usa o campo renomeado
	}, nil
}

// ============================================================================
// Lógica de Liderança (Ativo/Passivo)
// ============================================================================

type SimpleLeaderFollower struct {
	Queue *queue.QueueMaster
}
func (s *SimpleLeaderFollower) GetState() interface{} { return nil }
func (s *SimpleLeaderFollower) SetState(state []byte) error { return nil }
func (s *SimpleLeaderFollower) OnBecomeLeader() {
	log.Println("[Main] This node became the leader. Starting QueueMaster actor...")
	go s.Queue.Run()
}
func (s *SimpleLeaderFollower) OnBecomeFollower() {
	log.Println("[Main] This node became a follower. QueueMaster is idle.")
}

// ============================================================================
// Função Main
// ============================================================================
func main() {
	log.Println("Iniciando instância do serviço Jokenpo Queue...")

	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Fatal: Falha ao carregar configuração: %v", err)
	}
	log.Printf("[Main] Configuração carregada: ServiceName=%s, Port=%d, HealthPort=%d, ConsulAddrs=%s",
		cfg.ServiceName, cfg.ServicePort, cfg.HealthPort, cfg.ConsulAddrs)

	// --- MUDANÇA: Passa a lista de endereços para os construtores ---
	queueMaster := queue.NewQueueMaster(cfg.ConsulAddrs)
	elector, err := cluster.NewLeaderElector(cfg.ServiceName, cfg.ConsulAddrs)
	if err != nil {
		log.Fatalf("Fatal: Falha ao criar eleitor de líder: %v", err)
	}
	log.Println("[Main] Componentes QueueMaster e LeaderElector criados.")

	leaderFollowerHandler := &SimpleLeaderFollower{Queue: queueMaster}
	go elector.RunForLeadership(leaderFollowerHandler)
	log.Println("[Main] Campanha pela liderança iniciada em background.")

	mux := http.NewServeMux()
	mux.HandleFunc("/health", cluster.NewBasicHealthHandler())
	queue.RegisterQueueHandlers(mux, queueMaster, elector)
	log.Println("[Main] Handlers HTTP registrados para /queue/* e /health.")

	log.Println("[Main] Registrando serviço no Consul...")
	// --- MUDANÇA: Passa a lista de endereços para a função de registro ---
	err = cluster.RegisterServiceInConsul(cfg.ServiceName, cfg.ServicePort, cfg.HealthPort, cfg.ConsulAddrs)
	if err != nil {
		log.Fatalf("Fatal: Falha ao registrar serviço no Consul: %v", err)
	}

	listenAddress := fmt.Sprintf(":%d", cfg.ServicePort)
	log.Printf("[Main] Servidor HTTP do serviço Queue iniciando em %s.", listenAddress)

	if err := http.ListenAndServe(listenAddress, mux); err != nil {
		log.Fatalf("Fatal: Falha ao iniciar servidor HTTP: %v", err)
	}
}

//END OF FILE jokenpo/cmd/server/queue/main.go