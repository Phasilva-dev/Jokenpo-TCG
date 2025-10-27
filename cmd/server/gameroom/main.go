//START OF FILE jokenpo/cmd/server/gameroom/main.go
package main

import (
	"fmt"
	"jokenpo/internal/services/cluster"
	"jokenpo/internal/services/gameroom"
	"jokenpo/internal/services/api"
	"log"
	"net/http"
	"os"
	"strconv"
)

// ============================================================================
// Constantes de Configuração Padrão
// ============================================================================
const (
	defaultServiceName = "jokenpo-gameroom"
	defaultServicePort = 8083
	defaultHealthPort  = 8083
	// --- MUDANÇA: O padrão agora é uma lista de endereços ---
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
	serviceName := os.Getenv("GAMEROOM_SERVICE_NAME")
	if serviceName == "" {
		serviceName = defaultServiceName
	}

	// Lê a lista de endereços do Consul.
	consulAddrs := os.Getenv("CONSUL_HTTP_ADDR")
	if consulAddrs == "" {
		consulAddrs = defaultConsulAddr
	}

	servicePortStr := os.Getenv("GAMEROOM_SERVICE_PORT")
	if servicePortStr == "" {
		servicePortStr = fmt.Sprintf("%d", defaultServicePort)
	}
	servicePort, err := strconv.Atoi(servicePortStr)
	if err != nil {
		return nil, fmt.Errorf("formato de GAMEROOM_SERVICE_PORT inválido: %w", err)
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
// Função Main
// ============================================================================
func main() {
	log.Println("Iniciando instância do serviço Jokenpo GameRoom...")

	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Fatal: Falha ao carregar configuração: %v", err)
	}
	log.Printf("[Main] Configuração carregada: ServiceName=%s, Port=%d, HealthPort=%d, ConsulAddrs=%s",
		cfg.ServiceName, cfg.ServicePort, cfg.HealthPort, cfg.ConsulAddrs)

	roomManager := gameroom.NewRoomManager()
	go roomManager.Run()
	log.Println("[Main] RoomManager actor criado e iniciado.")

	mux := http.NewServeMux()
	mux.HandleFunc("/health", cluster.NewBasicHealthHandler())
	
	api.RegisterHandlers(mux, roomManager, cfg.ServicePort)
	log.Println("[Main] Handlers HTTP registrados para /rooms e /health.")

	log.Println("[Main] Registrando serviço no Consul...")
	// --- MUDANÇA: Passa a lista de endereços para a função de registro ---
	err = cluster.RegisterServiceInConsul(cfg.ServiceName, cfg.ServicePort, cfg.HealthPort, cfg.ConsulAddrs)
	if err != nil {
		log.Fatalf("Fatal: Falha ao registrar serviço no Consul: %v", err)
	}

	listenAddress := fmt.Sprintf(":%d", cfg.ServicePort)
	log.Printf("[Main] Servidor HTTP do serviço GameRoom iniciando em %s.", listenAddress)

	if err := http.ListenAndServe(listenAddress, mux); err != nil {
		log.Fatalf("Fatal: Falha ao iniciar servidor HTTP: %v", err)
	}
}
//END OF FILE jokenpo/cmd/server/gameroom/main.go