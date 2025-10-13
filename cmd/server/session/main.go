package main

import (
	"fmt"
	"jokenpo/internal/game/card"
	"jokenpo/internal/network"
	// --- MUDANÇA ---
	// O pacote de cluster agora é usado para registro e health check
	"jokenpo/internal/services/cluster"
	"jokenpo/internal/session"
	"log"
	"net/http" // Necessário para o novo handler
)

const (
	// --- MUDANÇA ---
	// Definimos o nome do serviço como uma constante para evitar erros de digitação.
	serviceName = "jokenpo-session"
	// A porta principal onde o serviço de WebSocket e o health check irão rodar.
	servicePort = 8080
)

func main() {
	// --- ETAPA 1: Lógica do Jogo (Inalterada) ---
	if err := card.InitGlobalCatalog(); err != nil {
		log.Fatalf("Falha fatal ao inicializar o catálogo de cartas: %v", err)
	}
	log.Println("Catálogo de cartas inicializado com sucesso.")

	gameHandler := session.NewGameHandler()
	go gameHandler.Matchmaker().Run()
	log.Println("Matchmaker iniciado.")

	server := network.NewServer(gameHandler)
	log.Println("Servidor de rede criado.")

	// --- ETAPA 2: CONFIGURAÇÃO DO CLUSTER E HEALTH CHECK (Atualizada) ---

	// 2.1 - Adiciona o handler de Health Check ao servidor HTTP principal.
	// O network.Server usa o http.DefaultServeMux por baixo dos panos,
	// então podemos registrar rotas nele antes de iniciá-lo.
	// O health check agora roda na mesma porta do serviço principal (8080).
	http.HandleFunc("/health", cluster.NewBasicHealthHandler()) // Usa o helper genérico
	log.Printf("Health Check handler registrado em :%d/health", servicePort)

	// 2.2 - Registra o serviço no Consul com o nome correto.
	// A porta de serviço e a porta de health check são agora a mesma.
	log.Printf("Registrando serviço '%s' no Consul...", serviceName)
	cluster.RegisterServiceInConsul(serviceName, servicePort, servicePort)

	// --- ETAPA 3: Iniciar Servidor Principal (Inalterada) ---
	address := fmt.Sprintf("0.0.0.0:%d", servicePort)
	log.Printf("Servidor principal (WebSocket & HTTP) iniciado em %s.", address)

	// A chamada server.Listen é bloqueante e agora servirá tanto o /ws quanto o /health.
	if err := server.Listen(address); err != nil {
		log.Fatalf("Falha fatal ao iniciar o servidor de rede: %v", err)
	}
}