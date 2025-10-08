// jokenpo/cmd/server/main.go
package main

import (
	"fmt"
	"log"
	"net/http" // Adicionado
	"os"      // Adicionado
	"strconv" // Adicionado

	"jokenpo/internal/game/card"
	"jokenpo/internal/network"
	"jokenpo/internal/session"

	consul "github.com/hashicorp/consul/api" // Adicionado
)

func main() {
	// --- ETAPA 1: Lógica do Jogo (Inalterada) ---
	if err := card.InitGlobalCatalog(); err != nil {
		log.Fatalf("Falha fatal ao inicializar o catálogo de cartas: %v", err)
	}
	fmt.Println("Catálogo de cartas inicializado com sucesso.")

	gameHandler := session.NewGameHandler()
	go gameHandler.Matchmaker().Run()
	fmt.Println("Matchmaker iniciado.")

	server := network.NewServer(gameHandler)
	fmt.Println("Servidor de rede criado.")

	// --- ETAPA 2: Adições para o Cluster ---

	// 2.1 - Inicia um servidor HTTP SÓ PARA HEALTH CHECK em outra goroutine
	healthPort := os.Getenv("HEALTH_CHECK_PORT") // Ex: "8000"
	if healthPort == "" {
		healthPort = "8000" // Porta padrão
	}
	go func() {
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "ok")
		})
		log.Printf("Servidor de Health Check escutando em :%s", healthPort)
		if err := http.ListenAndServe(":"+healthPort, nil); err != nil {
			log.Fatalf("Falha no servidor de Health Check: %v", err)
		}
	}()

	// 2.2 - Registra este serviço no Consul
	registerInConsul(healthPort)

	// --- ETAPA 3: Iniciar Servidor Principal (Inalterada) ---
	tcpAddress := "0.0.0.0:8080"
	go func() {
		if err := server.Listen(tcpAddress); err != nil {
			log.Fatalf("Falha fatal ao iniciar o servidor de rede TCP: %v", err)
		}
	}()
	fmt.Printf("Servidor principal TCP iniciado em %s.\n", tcpAddress)

	select {}
}

// Função para se registrar no Consul (copiada da nossa discussão anterior)
func registerInConsul(healthPort string) {
	config := consul.DefaultConfig()
    // O endereço do Consul é `consul:8500` dentro do Docker.
    // Para rodar localmente fora do Docker, seria `localhost:8500`.
	config.Address = os.Getenv("CONSUL_HTTP_ADDR")
	if config.Address == "" {
		config.Address = "consul:8500" // Padrão para ambiente Docker
	}

	consulClient, err := consul.NewClient(config)
	if err != nil {
		log.Fatalf("Erro ao criar cliente Consul: %s", err)
	}

    // Usamos o hostname do container como parte do ID para ser único
	serviceID := fmt.Sprintf("jokenpo-monolith-%s", os.Getenv("HOSTNAME"))
	serviceName := "jokenpo-monolith"
	
	// A porta que o Traefik vai usar é a do jogo (8080)
	servicePort := 8080 
	
	// O health check usa a porta que definimos
	healthPortInt, _ := strconv.Atoi(healthPort)

	registration := &consul.AgentServiceRegistration{
		ID:      serviceID,
		Name:    serviceName,
		Port:    servicePort,
		Address: os.Getenv("HOSTNAME"),
		Check: &consul.AgentServiceCheck{
			// O Consul vai chamar este endpoint dentro da rede do Docker
			HTTP:     fmt.Sprintf("http://%s:%d/health", os.Getenv("HOSTNAME"), healthPortInt),
			Interval: "10s",
			Timeout:  "1s",
		},
	}

	err = consulClient.Agent().ServiceRegister(registration)
	if err != nil {
		log.Fatalf("Falha ao registrar serviço no Consul: %s", err)
	}
	log.Printf("Serviço '%s' registrado no Consul com ID: %s", serviceName, serviceID)
}