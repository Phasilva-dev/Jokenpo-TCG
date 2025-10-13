package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	consul "github.com/hashicorp/consul/api"
)

// ============================================================================
// Constantes de Configuração Padrão (Sugestão Atendida)
// ============================================================================
// Estes são os valores padrão usados se as variáveis de ambiente não forem definidas.
// Mude aqui para alterar os padrões para o seu ambiente de desenvolvimento.
const (
	defaultListenAddr    = ":80"
	defaultConsulAddr    = "localhost:8500" // IP padrão definido como localhost
	defaultShutdownTimeout = 5 * time.Second
)


// ============================================================================
// Estruturas do Load Balancer (Inalteradas)
// ============================================================================

// BackendsStore armazena os backends disponíveis.
type BackendsStore struct {
	backends []*Backend
	mu       sync.RWMutex
	current  uint64
}

// Backend representa um servidor de backend com seu proxy reverso.
type Backend struct {
	URL          *url.URL
	ReverseProxy *httputil.ReverseProxy
}

func (s *BackendsStore) Set(newBackends []*Backend) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.backends = newBackends
}

func (s *BackendsStore) GetNext() *Backend {
	nextIndex := atomic.AddUint64(&s.current, uint64(1))
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.backends) == 0 {
		return nil
	}
	return s.backends[nextIndex%uint64(len(s.backends))]
}


// ============================================================================
// Lógica de Integração com o Consul (Inalterada)
// ============================================================================

func watchConsulServices(store *BackendsStore, consulClient *consul.Client, serviceName string) {
	// ... (código desta função permanece exatamente o mesmo) ...
	var waitIndex uint64 = 0

	for {
		opts := &consul.QueryOptions{WaitIndex: waitIndex}
		services, meta, err := consulClient.Health().Service(serviceName, "", true, opts)
		if err != nil {
			log.Printf("ERRO ao buscar serviço '%s' do Consul: %v.", serviceName, err)
			time.Sleep(5 * time.Second)
			continue
		}
		
		waitIndex = meta.LastIndex

		var newBackends []*Backend
		for _, service := range services {
			addr := service.Service.Address
			port := service.Service.Port
			backendURL, _ := url.Parse(fmt.Sprintf("http://%s:%d", addr, port))
			
			newBackends = append(newBackends, &Backend{
				URL:          backendURL,
				ReverseProxy: httputil.NewSingleHostReverseProxy(backendURL),
			})
		}
		
		if len(newBackends) > 0 {
			log.Printf("Configuração para '%s' atualizada: %d servidores saudáveis.", serviceName, len(newBackends))
			store.Set(newBackends)
		} else {
			log.Printf("AVISO: Nenhum backend saudável para '%s' encontrado.", serviceName)
			store.Set([]*Backend{})
		}
	}
}

func createConsulClient(addrs string) (*consul.Client, error) {
	// ... (código desta função permanece exatamente o mesmo) ...
	if addrs == "" {
		addrs = defaultConsulAddr
	}
	
	for _, addr := range strings.Split(addrs, ",") {
		config := consul.DefaultConfig()
		config.Address = strings.TrimSpace(addr)
		
		client, err := consul.NewClient(config)
		if err == nil {
			if _, err := client.Status().Leader(); err == nil {
				log.Printf("Conectado com sucesso ao agente Consul em %s", config.Address)
				return client, nil
			}
		}
		log.Printf("AVISO: Falha ao conectar ao agente Consul em %s: %v", config.Address, err)
	}
	
	return nil, fmt.Errorf("não foi possível conectar a nenhum dos agentes Consul fornecidos: %s", addrs)
}


// ============================================================================
// Lógica de Configuração (NOVO)
// ============================================================================

// Config armazena todas as configurações da aplicação.
type Config struct {
	ListenAddr      string
	ConsulAddrs     string
	TargetService   string
	ShutdownTimeout time.Duration
}

// loadConfig carrega a configuração a partir de variáveis de ambiente,
// usando as constantes como valores padrão.
func loadConfig() (*Config, error) {
	targetService := os.Getenv("LB_TARGET_SERVICE")
	if targetService == "" {
		return nil, fmt.Errorf("a variável de ambiente LB_TARGET_SERVICE deve ser definida")
	}

	consulAddrs := os.Getenv("CONSUL_ADDRS")
	if consulAddrs == "" {
		consulAddrs = defaultConsulAddr
	}

	listenAddr := os.Getenv("LB_LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = defaultListenAddr
	}
	
	return &Config{
		ListenAddr:      listenAddr,
		ConsulAddrs:     consulAddrs,
		TargetService:   targetService,
		ShutdownTimeout: defaultShutdownTimeout,
	}, nil
}


// ============================================================================
// Lógica dos Handlers HTTP (NOVO)
// ============================================================================

// newLoadBalancerHandler cria o handler principal para o proxy reverso.
func newLoadBalancerHandler(store *BackendsStore, targetService string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		backend := store.GetNext()
		if backend == nil {
			log.Printf("AVISO: Nenhum backend disponível para '%s'. Retornando 503.", targetService)
			http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
			return
		}
		log.Printf("LB para '%s': Redirecionando [%s] para [%s]", targetService, r.RemoteAddr, backend.URL)
		backend.ReverseProxy.ServeHTTP(w, r)
	}
}


// ============================================================================
// Função Main (Refatorada)
// ============================================================================
func main() {
	// 1. Carregar configuração
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("ERRO de configuração: %v", err)
	}

	// 2. Conectar ao Consul
	consulClient, err := createConsulClient(cfg.ConsulAddrs)
	if err != nil {
		log.Fatalf("Falha crítica ao conectar ao Consul: %v", err)
	}

	// 3. Inicializar o armazenamento de backends e o watcher do Consul
	store := &BackendsStore{}
	go watchConsulServices(store, consulClient, cfg.TargetService)

	// 4. Configurar o roteador HTTP com os handlers
	mux := http.NewServeMux()
	
	// Endpoint de Health Check para o próprio load balancer
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})

	// Handler principal do Load Balancer para todo o resto do tráfego
	mux.HandleFunc("/", newLoadBalancerHandler(store, cfg.TargetService))

	server := &http.Server{
		Addr:    cfg.ListenAddr,
		Handler: mux,
	}

	// 5. Iniciar o servidor e gerenciar o desligamento gracioso (graceful shutdown)
	go func() {
		log.Printf("Load Balancer para o serviço '%s' iniciado em '%s'", cfg.TargetService, cfg.ListenAddr)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Erro ao iniciar o servidor: %v", err)
		}
	}()
	
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Desligando o servidor...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Erro no desligamento gracioso: %v", err)
	}
	log.Println("Servidor desligado com sucesso.")
}