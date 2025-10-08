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
	"strings" // Importado para lidar com a lista de endereços
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	consul "github.com/hashicorp/consul/api"
)

// BackendsStore e Backend structs permanecem inalterados.
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

// watchConsulServices agora recebe o nome do serviço a ser monitorado.
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

// createConsulClient tenta conectar a uma lista de agentes do Consul.
func createConsulClient(addrs string) (*consul.Client, error) {
	if addrs == "" {
		addrs = "consul-1:8500" // Padrão se nada for fornecido
	}
	
	// Tenta cada endereço na lista até ter sucesso
	for _, addr := range strings.Split(addrs, ",") {
		config := consul.DefaultConfig()
		config.Address = strings.TrimSpace(addr)
		
		client, err := consul.NewClient(config)
		if err == nil {
			// Testa a conexão para garantir que o agente está acessível
			if _, err := client.Status().Leader(); err == nil {
				log.Printf("Conectado com sucesso ao agente Consul em %s", config.Address)
				return client, nil
			}
		}
		log.Printf("AVISO: Falha ao conectar ao agente Consul em %s: %v", config.Address, err)
	}
	
	return nil, fmt.Errorf("não foi possível conectar a nenhum dos agentes Consul fornecidos: %s", addrs)
}


func main() {
	// --- Configuração via Variáveis de Ambiente ---
	consulAddrs := os.Getenv("CONSUL_ADDRS") // Ex: "consul-1:8500,consul-2:8500"
	targetService := os.Getenv("LB_TARGET_SERVICE")
	if targetService == "" {
		log.Fatal("ERRO: A variável de ambiente LB_TARGET_SERVICE deve ser definida.")
	}

	consulClient, err := createConsulClient(consulAddrs)
	if err != nil {
		log.Fatalf("Falha crítica do Consul: %v", err)
	}

	store := &BackendsStore{}

	// Inicia o monitoramento do serviço alvo
	go watchConsulServices(store, consulClient, targetService)

	// O resto do código (servidor HTTP e graceful shutdown) permanece o mesmo.
	server := &http.Server{
		Addr: ":80",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			backend := store.GetNext()
			if backend == nil {
				http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
				return
			}
			log.Printf("LB para '%s': Redirecionando [%s] para [%s]", targetService, r.RemoteAddr, backend.URL)
			backend.ReverseProxy.ServeHTTP(w, r)
		}),
	}
	
	go func() {
		log.Printf("Load Balancer para o serviço '%s' iniciado na porta :80", targetService)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Erro ao iniciar o servidor: %v", err)
		}
	}()
	
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Desligando o servidor...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Erro no desligamento gracioso: %v", err)
	}
	log.Println("Servidor desligado com sucesso.")
}