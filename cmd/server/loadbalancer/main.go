//START OF FILE jokenpo/cmd/server/loadbalancer/main.go
package main

import (
	"context"
	"fmt"
	"jokenpo/internal/services/cluster" // <-- ADICIONA o import do pacote cluster
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	consul "github.com/hashicorp/consul/api"
)

// ============================================================================
// Constantes de Configuração
// ============================================================================
const (
	defaultListenAddr    = ":80"
	// O padrão agora é a lista completa, para alta disponibilidade.
	defaultConsulAddr    = "consul-1:8500,consul-2:8500,consul-3:8500"
	defaultShutdownTimeout = 5 * time.Second
)

// ============================================================================
// Estruturas do Load Balancer
// ============================================================================
// (Nenhuma mudança nesta seção)
type BackendsStore struct {
	backends []*Backend
	mu       sync.RWMutex
	current  uint64
}
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
// Lógica de Integração com o Consul
// ============================================================================
// (Nenhuma mudança nesta seção)
func watchConsulServices(store *BackendsStore, consulClient *consul.Client, serviceName string) {
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

// --- REMOVIDO: A função createConsulClient foi movida para o pacote cluster ---

// ============================================================================
// Lógica de Configuração
// ============================================================================
type Config struct {
	ListenAddr      string
	ConsulAddrs     string
	TargetService   string
	ShutdownTimeout time.Duration
}

func loadConfig() (*Config, error) {
	targetService := os.Getenv("LB_TARGET_SERVICE")
	if targetService == "" {
		return nil, fmt.Errorf("a variável de ambiente LB_TARGET_SERVICE deve ser definida")
	}

	// --- MUDANÇA: Usa a variável de ambiente padrão CONSUL_HTTP_ADDR ---
	consulAddrs := os.Getenv("CONSUL_HTTP_ADDR")
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
// Lógica dos Handlers HTTP
// ============================================================================
// (Nenhuma mudança nesta seção)
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
// Função Main
// ============================================================================
func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("ERRO de configuração: %v", err)
	}

	// --- MUDANÇA: Usa a função helper centralizada do pacote cluster ---
	consulClient, err := cluster.NewConsulClient(cfg.ConsulAddrs)
	if err != nil {
		log.Fatalf("Falha crítica ao conectar ao Consul: %v", err)
	}

	store := &BackendsStore{}
	go watchConsulServices(store, consulClient, cfg.TargetService)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK); fmt.Fprintln(w, "OK")
	})
	mux.HandleFunc("/", newLoadBalancerHandler(store, cfg.TargetService))

	server := &http.Server{Addr: cfg.ListenAddr, Handler: mux}

	go func() {
		log.Printf("Load Balancer para '%s' iniciado em '%s'", cfg.TargetService, cfg.ListenAddr)
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

//END OF FILE jokenpo/cmd/server/loadbalancer/main.go