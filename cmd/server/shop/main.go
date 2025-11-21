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
	if serviceName == "" { serviceName = defaultServiceName }
	consulAddrs := os.Getenv("CONSUL_HTTP_ADDR")
	if consulAddrs == "" { consulAddrs = defaultConsulAddr }
	servicePortStr := os.Getenv("SHOP_SERVICE_PORT")
	if servicePortStr == "" { servicePortStr = fmt.Sprintf("%d", defaultServicePort) }
	servicePort, err := strconv.Atoi(servicePortStr)
	if err != nil { return nil, fmt.Errorf("port invalid: %w", err) }
	healthPortStr := os.Getenv("HEALTH_CHECK_PORT")
	if healthPortStr == "" { healthPortStr = fmt.Sprintf("%d", defaultHealthPort) }
	healthPort, err := strconv.Atoi(healthPortStr)
	if err != nil { return nil, fmt.Errorf("health port invalid: %w", err) }
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
	if err != nil { log.Fatalf("Fatal: %v", err) }

	consulManager, err := cluster.NewConsulManager(cfg.ConsulAddrs)
	if err != nil { log.Fatalf("Fatal: %v", err) }

	advertisedHost := os.Getenv("SERVICE_ADVERTISED_HOSTNAME")
	if advertisedHost == "" {
		hostname, _ := os.Hostname()
		advertisedHost = hostname
	}
	registrar, err := cluster.NewServiceRegistrar(
		consulManager,
		cfg.ServiceName,
		advertisedHost,
		cfg.ServicePort,
		cfg.HealthPort,
	)
	if err != nil { log.Fatalf("Fatal: %v", err) }

	consulManager.OnReconnect(registrar.Register)
	registrar.Register()

    // --- MUDANÇA: Passa o consulManager para o serviço ---
	shopService := shop.NewShopService(consulManager)
	log.Println("[Main] Ator do ShopService criado.")

	elector, err := cluster.NewLeaderElector(cfg.ServiceName, consulManager, advertisedHost)
	if err != nil { log.Fatalf("Fatal: %v", err) }

	go elector.RunForLeadership(shopService)

	shopHandler := shop.CreateShopHandler(shopService, elector)
	http.HandleFunc("/health", cluster.NewBasicHealthHandler())
	http.HandleFunc("/Purchase", shopHandler)

	listenAddress := fmt.Sprintf(":%d", cfg.ServicePort)
	log.Printf("[Main] Servidor HTTP iniciando em %s.", listenAddress)

	if err := http.ListenAndServe(listenAddress, nil); err != nil {
		log.Fatalf("Fatal: %v", err)
	}
}
//END OF FILE jokenpo/cmd/server/shop/main.go