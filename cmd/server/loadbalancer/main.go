//START OF FILE jokenpo/cmd/server/loadbalancer/main.go
package main

import (
	"context"
	"fmt"
	"jokenpo/internal/services/cluster"
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

const (
	defaultListenAddr      = ":80"
	defaultConsulAddr      = "consul-1:8500,consul-2:8500,consul-3:8500"
	defaultShutdownTimeout = 5 * time.Second
)

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

func watchConsulServices(store *BackendsStore, manager *cluster.ConsulManager, serviceName string) {
	var waitIndex uint64 = 0
	for {
		client := manager.GetClient()
		if client == nil {
			log.Printf("[Watcher] AVISO: Cliente Consul não está disponível. Tentando novamente em 5s.")
			time.Sleep(5 * time.Second)
			continue
		}

		opts := &consul.QueryOptions{WaitIndex: waitIndex, WaitTime: 2 * time.Minute}
		services, meta, err := client.Health().Service(serviceName, "", true, opts)

		if err != nil {
			log.Printf("[Watcher] ERRO ao buscar serviço '%s' do Consul: %v. O manager tentará reconectar.", serviceName, err)
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

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("ERRO de configuração: %v", err)
	}

	consulManager, err := cluster.NewConsulManager(cfg.ConsulAddrs)
	if err != nil {
		log.Fatalf("Falha crítica ao criar Consul Manager: %v", err)
	}

	store := &BackendsStore{}
	go watchConsulServices(store, consulManager, cfg.TargetService)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
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