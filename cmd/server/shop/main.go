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
	defaultConsulAddr  = "consul-1:8500,consul-2:8500,consul-3:8500" // Atualizado para lista
)

type Config struct {
	ServiceName string
	ServicePort int
	HealthPort  int
	ConsulAddrs string // Renomeado de ConsulAddr para ConsulAddrs
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

	// --- MUDANÇA: Cria o ConsulManager uma vez ---
	consulManager, err := cluster.NewConsulManager(cfg.ConsulAddrs)
	if err != nil {
		log.Fatalf("Fatal: Falha ao criar Consul Manager: %v", err)
	}

	shopService := shop.NewShopService()
	log.Println("[Main] Ator do ShopService criado.")

	// --- MUDANÇA: Passa o manager para o eleitor de líder ---
	elector, err := cluster.NewLeaderElector(cfg.ServiceName, consulManager)
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

	err = cluster.RegisterServiceInConsul(cfg.ServiceName, cfg.ServicePort, cfg.HealthPort, cfg.ConsulAddrs)
	if err != nil {
		log.Fatalf("Fatal: Falha ao registrar serviço no Consul: %v", err)
	}

	listenAddress := fmt.Sprintf(":%d", cfg.ServicePort)
	log.Printf("[Main] Servidor HTTP do serviço Shop iniciando em %s.", listenAddress)

	if err := http.ListenAndServe(listenAddress, nil); err != nil {
		log.Fatalf("Fatal: Falha ao iniciar servidor HTTP: %v", err)
	}
}

//END OF FILE jokenpo/cmd/server/shop/main.go