package main

import (
	"fmt"
	"jokenpo/internal/game/card"
	"jokenpo/internal/network"
	"jokenpo/internal/services/cluster"
	"jokenpo/internal/session"
	"log"
	"net/http"
	"os"
	"strconv"
)

// ============================================================================
// Constantes de Configuração Padrão
// ============================================================================
const (
	defaultServiceName = "jokenpo-session"
	defaultServicePort = 8080
	defaultHealthPort  = 8080 // Por padrão, a mesma porta do serviço
	defaultConsulAddr  = "consul-1:8500"
)

// ============================================================================
// Lógica de Configuração
// ============================================================================

// Config armazena todas as configurações da aplicação.
type Config struct {
	ServiceName string
	ServicePort int
	HealthPort  int
	ConsulAddr  string
}

// loadConfig carrega a configuração a partir de variáveis de ambiente.
func loadConfig() (*Config, error) {
	serviceName := os.Getenv("SESSION_SERVICE_NAME")
	if serviceName == "" {
		serviceName = defaultServiceName
	}

	consulAddr := os.Getenv("CONSUL_HTTP_ADDR")
	if consulAddr == "" {
		consulAddr = defaultConsulAddr
	}

	servicePortStr := os.Getenv("SESSION_SERVICE_PORT")
	if servicePortStr == "" {
		servicePortStr = fmt.Sprintf("%d", defaultServicePort)
	}
	servicePort, err := strconv.Atoi(servicePortStr)
	if err != nil {
		return nil, fmt.Errorf("formato de SESSION_SERVICE_PORT inválido: %w", err)
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
// Função Main (Refatorada)
// ============================================================================
func main() {
	// 1. CARREGA A CONFIGURAÇÃO
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Fatal: Falha ao carregar configuração: %v", err)
	}
	log.Printf("[Main] Configuração carregada: ServiceName=%s, Port=%d, HealthPort=%d, Consul=%s",
		cfg.ServiceName, cfg.ServicePort, cfg.HealthPort, cfg.ConsulAddr)

	// 2. INICIA A LÓGICA DO JOGO
	if err := card.InitGlobalCatalog(); err != nil {
		log.Fatalf("Falha fatal ao inicializar o catálogo de cartas: %v", err)
	}
	log.Println("[Main] Catálogo de cartas inicializado com sucesso.")

	gameHandler, err := session.NewGameHandler(cfg.ConsulAddr)
	if err != nil {
		log.Fatalf("Initial Catalog failure: %s",err)
	}
	go gameHandler.Matchmaker().Run()
	log.Println("[Main] Matchmaker iniciado.")

	server := network.NewServer(gameHandler)
	log.Println("[Main] Servidor de rede criado.")

	// 3. CONFIGURA O CLUSTER E O HEALTH CHECK
	http.HandleFunc("/health", cluster.NewBasicHealthHandler())
	log.Printf("[Main] Health Check handler registrado em :%d/health", cfg.ServicePort)

	log.Printf("[Main] Registrando serviço '%s' no Consul...", cfg.ServiceName)
	err = cluster.RegisterServiceInConsul(cfg.ServiceName, cfg.ServicePort, cfg.HealthPort, cfg.ConsulAddr)
	if err != nil {
		log.Fatalf("Fatal: Falha ao registrar serviço no Consul: %v", err)
	}

	// 4. INICIA O SERVIDOR PRINCIPAL
	address := fmt.Sprintf("0.0.0.0:%d", cfg.ServicePort)
	log.Printf("[Main] Servidor principal (WebSocket & HTTP) iniciado em %s.", address)

	if err := server.Listen(address); err != nil {
		log.Fatalf("Falha fatal ao iniciar o servidor de rede: %v", err)
	}
}