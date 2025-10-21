//START OF FILE jokenpo/cmd/server/queue/main.go
package main

import (
	"fmt"
	"jokenpo/internal/api" // Importa o pacote com os handlers da API
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
	defaultServicePort = 8082 // Porta dedicada para este serviço
	defaultHealthPort  = 8082
	defaultConsulAddr  = "consul-1:8500"
)

// ============================================================================
// Estrutura de Configuração
// ============================================================================
type Config struct {
	ServiceName string
	ServicePort int
	HealthPort  int
	ConsulAddr  string
}

// loadConfig carrega a configuração a partir de variáveis de ambiente.
func loadConfig() (*Config, error) {
	serviceName := os.Getenv("QUEUE_SERVICE_NAME")
	if serviceName == "" {
		serviceName = defaultServiceName
	}

	consulAddr := os.Getenv("CONSUL_HTTP_ADDR")
	if consulAddr == "" {
		consulAddr = defaultConsulAddr
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
		ConsulAddr:  consulAddr,
	}, nil
}

// ============================================================================
// Lógica de Liderança (Ativo/Passivo)
// ============================================================================

// SimpleLeaderFollower implementa a interface StatefulService para controlar
// o ciclo de vida do QueueMaster.
type SimpleLeaderFollower struct {
	Queue *queue.QueueMaster
}
// Não persistimos o estado das filas. Se o líder cair, as filas são resetadas.
func (s *SimpleLeaderFollower) GetState() interface{} { return nil }
func (s *SimpleLeaderFollower) SetState(state []byte) error { return nil }

// OnBecomeLeader é o callback crucial que inicia o ator do QueueMaster.
func (s *SimpleLeaderFollower) OnBecomeLeader() {
	log.Println("[Main] This node became the leader. Starting QueueMaster actor...")
	// SÓ o líder executa a lógica de pareamento.
	go s.Queue.Run()
}

// OnBecomeFollower apenas loga a mudança de estado.
func (s *SimpleLeaderFollower) OnBecomeFollower() {
	log.Println("[Main] This node became a follower. QueueMaster is idle.")
	// A goroutine Run() do líder antigo morrerá junto com o processo dele.
	// O novo seguidor não inicia uma nova goroutine.
}

// ============================================================================
// Função Main
// ============================================================================
func main() {
	log.Println("Iniciando instância do serviço Jokenpo Queue...")

	// 1. CARREGA A CONFIGURAÇÃO
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Fatal: Falha ao carregar configuração: %v", err)
	}
	log.Printf("[Main] Configuração carregada: ServiceName=%s, Port=%d, HealthPort=%d, Consul=%s",
		cfg.ServiceName, cfg.ServicePort, cfg.HealthPort, cfg.ConsulAddr)

	// 2. CRIA AS INSTÂNCIAS DOS COMPONENTES
	queueMaster := queue.NewQueueMaster()
	elector, err := cluster.NewLeaderElector(cfg.ServiceName, cfg.ConsulAddr)
	if err != nil {
		log.Fatalf("Fatal: Falha ao criar eleitor de líder: %v", err)
	}
	log.Println("[Main] Componentes QueueMaster e LeaderElector criados.")

	// 3. INICIA A CAMPANHA PELA LIDERANÇA
	// O leaderFollowerHandler garante que o queueMaster.Run() só seja chamado no líder.
	leaderFollowerHandler := &SimpleLeaderFollower{Queue: queueMaster}
	go elector.RunForLeadership(leaderFollowerHandler)
	log.Println("[Main] Campanha pela liderança iniciada em background.")

	// 4. CONFIGURA OS HANDLERS DA API HTTP
	mux := http.NewServeMux()
	mux.HandleFunc("/health", cluster.NewBasicHealthHandler())
	api.RegisterQueueHandlers(mux, queueMaster, elector) // Registra todas as rotas /queue/*
	log.Println("[Main] Handlers HTTP registrados para /queue/* e /health.")

	// 5. REGISTRA O SERVIÇO NO CONSUL
	log.Println("[Main] Registrando serviço no Consul...")
	err = cluster.RegisterServiceInConsul(cfg.ServiceName, cfg.ServicePort, cfg.HealthPort, cfg.ConsulAddr)
	if err != nil {
		log.Fatalf("Fatal: Falha ao registrar serviço no Consul: %v", err)
	}

	// 6. INICIA O SERVIDOR HTTP
	listenAddress := fmt.Sprintf(":%d", cfg.ServicePort)
	log.Printf("[Main] Servidor HTTP do serviço Queue iniciando em %s.", listenAddress)

	if err := http.ListenAndServe(listenAddress, mux); err != nil {
		log.Fatalf("Fatal: Falha ao iniciar servidor HTTP: %v", err)
	}
}

//END OF FILE jokenpo/cmd/server/queue/main.go