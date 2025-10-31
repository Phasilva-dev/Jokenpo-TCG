//START OF FILE jokenpo/cmd/server/shop/main.go
package main

import (
	"fmt"
	"jokenpo/internal/services/cluster"
	"jokenpo/internal/services/shop"
	"log"
	"net/http"
	"os"
	"strconv"
)

const (
	defaultServiceName = "jokenpo-shop"
	defaultServicePort = 8081
	defaultHealthPort  = 8081
	defaultConsulAddr  = "consul-1:8500,consul-2:8500,consul-3:8500"
)

type Config struct {
	ServiceName string
	ServicePort int
	HealthPort  int
	ConsulAddrs string
}

func loadConfig() (*Config, error) {
	serviceName := os.Getenv("SHOP_SERVICE_NAME")
	if serviceName == "" {
		serviceName = defaultServiceName
	}
	consulAddrs := os.Getenv("CONSUL_HTTP_ADDR")
	if consulAddrs == "" {
		consulAddrs = defaultConsulAddr
	}
	servicePortStr := os.Getenv("SHOP_SERVICE_PORT")
	if servicePortStr == "" {
		servicePortStr = fmt.Sprintf("%d", defaultServicePort)
	}
	servicePort, err := strconv.Atoi(servicePortStr)
	if err != nil {
		return nil, fmt.Errorf("formato de SHOP_SERVICE_PORT inválido: %w", err)
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

func main() {
	log.Println("Iniciando instância do serviço Jokenpo Shop...")

	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Fatal: Falha ao carregar configuração: %v", err)
	}
	log.Printf("[Main] Configuração carregada: ServiceName=%s, Port=%d, HealthPort=%d, ConsulAddrs=%s",
		cfg.ServiceName, cfg.ServicePort, cfg.HealthPort, cfg.ConsulAddrs)

	// --- LÓGICA DE REGISTRO RESILIENTE ---
	// 1. Cria o ConsulManager, que gerencia a conexão de forma contínua.
	consulManager, err := cluster.NewConsulManager(cfg.ConsulAddrs)
	if err != nil {
		log.Fatalf("Fatal: Falha ao criar Consul Manager: %v", err)
	}

	// 2. Cria o ServiceRegistrar, que sabe como registrar este serviço.
	advertisedHost := os.Getenv("SERVICE_ADVERTISED_HOSTNAME")
	if advertisedHost == "" {
		// Fallback: se a variável não for definida, usa o hostname do próprio contêiner.
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

	// 3. Conecta os dois: toda vez que o manager se reconectar, ele tentará registrar o serviço novamente.
	consulManager.OnReconnect(registrar.Register)

	// 4. CORREÇÃO: Realiza o primeiro registro manualmente na inicialização.
	registrar.Register()
	// --- FIM DA LÓGICA DE REGISTRO RESILIENTE ---

	shopService := shop.NewShopService()
	log.Println("[Main] Ator do ShopService criado.")

	elector, err := cluster.NewLeaderElector(cfg.ServiceName, consulManager, advertisedHost)
	if err != nil {
		log.Fatalf("Fatal: Falha ao criar eleitor de líder: %v", err)
	}
	log.Println("[Main] LeaderElector criado.")

	go elector.RunForLeadership(shopService)
	log.Println("[Main] Campanha pela liderança iniciada em background.")

	shopHandler := shop.CreateShopHandler(shopService, elector)
	http.HandleFunc("/health", cluster.NewBasicHealthHandler())
	http.HandleFunc("/Purchase", shopHandler)
	log.Println("[Main] Handlers HTTP registrados.")

	// A chamada antiga e única ao RegisterServiceInConsul foi removida.
	// O gerenciamento agora é contínuo.

	listenAddress := fmt.Sprintf(":%d", cfg.ServicePort)
	log.Printf("[Main] Servidor HTTP do serviço Shop iniciando em %s.", listenAddress)

	if err := http.ListenAndServe(listenAddress, nil); err != nil {
		log.Fatalf("Fatal: Falha ao iniciar servidor HTTP: %v", err)
	}
}

//END OF FILE jokenpo/cmd/server/shop/main.go