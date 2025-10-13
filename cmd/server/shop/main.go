package main

import (
	"fmt"
	"jokenpo/internal/api"
	"jokenpo/internal/services/cluster"
	"jokenpo/internal/services/shop"
	"log"
	"net/http"
	"os"
	"strconv"
)

// ============================================================================
// Constantes de Configuração Padrão
// ============================================================================
const (
	defaultServiceName = "jokenpo-shop"
	defaultServicePort = 8081
	// --- MUDANÇA ---
	// Adicionamos uma constante para a porta de health check. Por padrão, é a mesma.
	defaultHealthPort = 8081
	defaultConsulAddr = "consul-1:8500"
)

// ============================================================================
// Lógica de Configuração
// ============================================================================

// --- MUDANÇA ---
// Config agora tem um campo explícito para HealthPort.
type Config struct {
	ServiceName string
	ServicePort int
	HealthPort  int // Novo campo
	ConsulAddr  string
}

// loadConfig carrega a configuração a partir de variáveis de ambiente.
func loadConfig() (*Config, error) {
	serviceName := os.Getenv("SHOP_SERVICE_NAME")
	if serviceName == "" {
		serviceName = defaultServiceName
	}

	consulAddr := os.Getenv("CONSUL_HTTP_ADDR")
	if consulAddr == "" {
		consulAddr = defaultConsulAddr
	}

	servicePortStr := os.Getenv("SHOP_SERVICE_PORT")
	if servicePortStr == "" {
		servicePortStr = fmt.Sprintf("%d", defaultServicePort)
	}
	servicePort, err := strconv.Atoi(servicePortStr)
	if err != nil {
		return nil, fmt.Errorf("formato de SHOP_SERVICE_PORT inválido: %w", err)
	}

	// --- MUDANÇA ---
	// Carrega a porta de health check de sua própria variável de ambiente.
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
		HealthPort:  healthPort, // Adiciona ao struct
		ConsulAddr:  consulAddr,
	}, nil
}


// ============================================================================
// Função Main (Refatorada)
// ============================================================================
func main() {
	log.Println("Iniciando instância do serviço Jokenpo Shop...")

	// 1. CARREGA A CONFIGURAÇÃO
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Fatal: Falha ao carregar configuração: %v", err)
	}
	log.Printf("[Main] Configuração carregada: ServiceName=%s, Port=%d, HealthPort=%d, Consul=%s",
		cfg.ServiceName, cfg.ServicePort, cfg.HealthPort, cfg.ConsulAddr)

	// 2. CRIA A INSTÂNCIA DO SERVIÇO DE NEGÓCIO
	shopService := shop.NewShopService()
	log.Println("[Main] Ator do ShopService criado.")

	// 3. CRIA O ELEITOR DE LÍDER
	elector, err := cluster.NewLeaderElector(cfg.ServiceName, cfg.ConsulAddr)
	if err != nil {
		log.Fatalf("Fatal: Falha ao criar eleitor de líder: %v", err)
	}
	log.Println("[Main] LeaderElector criado.")

	// 4. INICIA A CAMPANHA PELA LIDERANÇA
	go elector.RunForLeadership(shopService)
	log.Println("[Main] Campanha pela liderança iniciada em background.")

	// 5. CONFIGURA OS HANDLERS DA API HTTP
	shopHandler := api.CreateShopHandler(shopService, elector)
	http.HandleFunc("/health", cluster.NewBasicHealthHandler())
	http.HandleFunc("/Purchase", shopHandler)
	log.Println("[Main] Handlers HTTP registrados.")

	// 6. REGISTRA O SERVIÇO NO CONSUL
	log.Println("[Main] Registrando serviço no Consul...")
	// --- MUDANÇA ---
	// A chamada agora é explícita e usa o campo HealthPort da configuração.
	err = cluster.RegisterServiceInConsul(cfg.ServiceName, cfg.ServicePort, cfg.HealthPort, cfg.ConsulAddr)
	if err != nil {
		log.Fatalf("Fatal: Falha ao registrar serviço no Consul: %v", err)
	}

	// 7. INICIA O SERVIDOR
	listenAddress := fmt.Sprintf(":%d", cfg.ServicePort)
	log.Printf("[Main] Servidor HTTP do serviço Shop iniciando em %s.", listenAddress)

	if err := http.ListenAndServe(listenAddress, nil); err != nil {
		log.Fatalf("Fatal: Falha ao iniciar servidor HTTP: %v", err)
	}
}