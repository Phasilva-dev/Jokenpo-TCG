//START OF FILE jokenpo/cmd/server/session/main.go
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

const (
	defaultServiceName = "jokenpo-session"
	defaultServicePort = 8080
	defaultHealthPort  = 8080
	defaultConsulAddr  = "consul-1:8500,consul-2:8500,consul-3:8500"
)

type Config struct {
	ServiceName        string
	ServicePort        int
	HealthPort         int
	ConsulAddrs        string
	AdvertisedHostname string
}

func loadConfig() (*Config, error) {
	serviceName := os.Getenv("SESSION_SERVICE_NAME")
	if serviceName == "" {
		serviceName = defaultServiceName
	}
	consulAddrs := os.Getenv("CONSUL_HTTP_ADDR")
	if consulAddrs == "" {
		consulAddrs = defaultConsulAddr
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
	advertisedHostname := os.Getenv("SERVICE_ADVERTISED_HOSTNAME")
	if advertisedHostname == "" {
		// Fallback: usa o hostname do próprio container se a variável não estiver definida
		hostname, err := os.Hostname()
		if err != nil {
			return nil, fmt.Errorf("falha ao obter hostname do container: %w", err)
		}
		advertisedHostname = hostname
	}
	return &Config{
		ServiceName:        serviceName,
		ServicePort:        servicePort,
		HealthPort:         healthPort,
		ConsulAddrs:        consulAddrs,
		AdvertisedHostname: advertisedHostname,
	}, nil
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Fatal: Falha ao carregar configuração: %v", err)
	}
	log.Printf("[Main] Configuração carregada: ServiceName=%s, Port=%d, HealthPort=%d, ConsulAddrs=%s, AdvertiseHost=%s",
		cfg.ServiceName, cfg.ServicePort, cfg.HealthPort, cfg.ConsulAddrs, cfg.AdvertisedHostname)

	if err := card.InitGlobalCatalog(); err != nil {
		log.Fatalf("Falha fatal ao inicializar o catálogo de cartas: %v", err)
	}
	log.Println("[Main] Catálogo de cartas inicializado com sucesso.")

	// --- LÓGICA DE REGISTRO RESILIENTE ---
	// 1. Cria o ConsulManager, que gerencia a conexão de forma contínua.
	consulManager, err := cluster.NewConsulManager(cfg.ConsulAddrs)
	if err != nil {
		log.Fatalf("Fatal: Falha ao criar Consul Manager: %v", err)
	}

	// 2. Cria o ServiceRegistrar, que sabe como registrar este serviço.
	registrar, err := cluster.NewServiceRegistrar(
		consulManager,
		cfg.ServiceName,
		cfg.AdvertisedHostname,
		cfg.ServicePort,
		cfg.HealthPort,
	)
	if err != nil {
		log.Fatalf("Fatal: Falha ao criar o Service Registrar: %v", err)
	}

	// 3. Conecta os dois: toda vez que o manager se reconectar, ele tentará registrar o serviço novamente.
	consulManager.OnReconnect(registrar.Register)

	// 4. CORREÇÃO: Realiza o primeiro registro manualmente na inicialização.
	registrar.Register()
	// --- FIM DA LÓGICA DE REGISTRO RESILIENTE ---

	gameHandler, err := session.NewGameHandler(consulManager, cfg.AdvertisedHostname)
	if err != nil {
		log.Fatalf("Falha ao criar o GameHandler: %v", err)
	}
	log.Println("[Main] GameHandler criado.")

	server := network.NewServer(gameHandler)
	log.Println("[Main] Servidor de rede criado.")

	http.HandleFunc("/health", cluster.NewBasicHealthHandler())
	http.HandleFunc("/match-found", gameHandler.CallbackMatchFound)
	http.HandleFunc("/trade-found", gameHandler.CallbackTradeFound)
	http.HandleFunc("/game-event", gameHandler.CallbackGameEvent)
	log.Printf("[Main] Handlers de Health Check e Callback registrados.")

	// A chamada antiga e única ao RegisterServiceInConsul foi removida.

	address := fmt.Sprintf("0.0.0.0:%d", cfg.ServicePort)
	log.Printf("[Main] Servidor principal (WebSocket & HTTP) iniciado em %s.", address)

	if err := server.Listen(address); err != nil {
		log.Fatalf("Falha fatal ao iniciar o servidor de rede: %v", err)
	}
}
//END OF FILE jokenpo/cmd/server/session/main.go