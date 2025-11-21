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

const (
	defaultServiceName = "jokenpo-queue"
	defaultServicePort = 8082
	defaultHealthPort  = 8082
	defaultConsulAddr  = "consul-1:8500,consul-2:8500,consul-3:8500"
)

type Config struct {
	ServiceName string
	ServicePort int
	HealthPort  int
	ConsulAddrs string
}

func loadConfig() (*Config, error) {
	serviceName := os.Getenv("QUEUE_SERVICE_NAME")
	if serviceName == "" {
		serviceName = defaultServiceName
	}
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
		ConsulAddrs: consulAddrs,
	}, nil
}

type SimpleLeaderFollower struct {
	Queue *queue.QueueMaster
}

func (s *SimpleLeaderFollower) GetState() interface{}      { return nil }
func (s *SimpleLeaderFollower) SetState(state []byte) error { return nil }
func (s *SimpleLeaderFollower) OnBecomeLeader() {
	log.Println("[Main] This node became the leader. Starting QueueMaster actor...")
	go s.Queue.Run()
}
func (s *SimpleLeaderFollower) OnBecomeFollower() {
	log.Println("[Main] This node became a follower. QueueMaster is idle.")
}

func main() {
	log.Println("Iniciando instância do serviço Jokenpo Queue...")

	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Fatal: Falha ao carregar configuração: %v", err)
	}
	log.Printf("[Main] Configuração carregada: ServiceName=%s, Port=%d, HealthPort=%d, ConsulAddrs=%s",
		cfg.ServiceName, cfg.ServicePort, cfg.HealthPort, cfg.ConsulAddrs)

	// 1. Cria o ConsulManager
	consulManager, err := cluster.NewConsulManager(cfg.ConsulAddrs)
	if err != nil {
		log.Fatalf("Fatal: Falha ao criar Consul Manager: %v", err)
	}

	advertisedHost := os.Getenv("SERVICE_ADVERTISED_HOSTNAME")
	if advertisedHost == "" {
		hostname, err := os.Hostname()
		if err != nil {
			log.Fatalf("Fatal: Falha ao obter hostname do contêiner: %v", err)
		}
		advertisedHost = hostname
	}
	registrar, err := cluster.NewServiceRegistrar(
		consulManager,
		cfg.ServiceName,
		advertisedHost,
		cfg.ServicePort,
		cfg.HealthPort,
	)
	if err != nil {
		log.Fatalf("Fatal: Falha ao criar o Service Registrar: %v", err)
	}

	consulManager.OnReconnect(registrar.Register)
	registrar.Register()

    // --- MUDANÇA CRUCIAL AQUI ---
    // Passamos o consulManager para o QueueMaster poder descobrir a Blockchain
	queueMaster := queue.NewQueueMaster(consulManager)
	
    elector, err := cluster.NewLeaderElector(cfg.ServiceName, consulManager, advertisedHost)
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

	listenAddress := fmt.Sprintf(":%d", cfg.ServicePort)
	log.Printf("[Main] Servidor HTTP do serviço Queue iniciando em %s.", listenAddress)

	if err := http.ListenAndServe(listenAddress, mux); err != nil {
		log.Fatalf("Fatal: Falha ao iniciar servidor HTTP: %v", err)
	}
}
//END OF FILE jokenpo/cmd/server/queue/main.go