package main

import (
	"fmt"
	// SEU PACOTE API PARECE SER GENÉRICO, VAMOS ASSUMIR QUE O HANDLER DO SHOP ESTÁ EM SEU PRÓPRIO PACOTE
	"jokenpo/internal/api"
	"jokenpo/internal/services/cluster" // Assumindo que RegisterServiceInConsul, NewBasicHealthHandler, NewLeaderElector estão aqui
	"jokenpo/internal/services/shop"
	"log"
	"net/http"
)

// --- Configuração do Serviço ---
const (
	// O nome que o Broker e o LeaderElector usarão para identificar este cluster.
	serviceName = "jokenpo-shop"
	// A porta onde este serviço irá escutar por requisições HTTP.
	servicePort = 8081
)

func main() {
	log.Println("Starting Jokenpo Shop Service instance...")

	// 1. CRIA A INSTÂNCIA DO SERVIÇO DE NEGÓCIO
	// Esta é a implementação concreta que gerenciará o estado do shop.
	shopService := shop.NewShopService()
	log.Println("[Main] ShopService actor created.")

	// 2. CRIA O ELEITOR DE LÍDER
	// Ele será responsável por coordenar com o Consul para decidir quem é o líder.
	elector, err := cluster.NewLeaderElector(serviceName)
	if err != nil {
		log.Fatalf("Fatal: Failed to create leader elector: %v", err)
	}
	log.Println("[Main] LeaderElector created.")

	// 3. INICIA A CAMPANHA PELA LIDERANÇA
	// Executamos isso em uma goroutine para não bloquear o início do servidor HTTP.
	// O eleitor agora gerencia o ciclo de vida do shopService (chamando OnBecomeLeader, etc.).
	go elector.RunForLeadership(shopService)
	log.Println("[Main] Leadership campaign started in the background.")

	// 4. CONFIGURA OS HANDLERS DA API HTTP
	// Criamos o handler principal, passando tanto o serviço (para fazer o trabalho)
	// quanto o eleitor (para verificar a permissão para fazer o trabalho).
	shopHandler := api.CreateShopHandler(shopService, elector)

	// O health check que o Consul usará deve ser um "liveness check" simples.
	// Ele apenas verifica se o processo e o servidor HTTP estão respondendo.
	// A lógica de "readiness" (estou pronto para receber tráfego?) é tratada
	// pela verificação `!elector.IsLeader()` dentro do shopHandler.
	http.HandleFunc("/health", cluster.NewBasicHealthHandler()) // Assumindo que este helper existe no pacote cluster
	http.HandleFunc("/Purchase", shopHandler)
	log.Println("[Main] HTTP handlers registered.")

	// 5. REGISTRA O SERVIÇO NO CONSUL
	// Anuncia a existência desta instância para o Broker.
	log.Println("[Main] Registering service with Consul...")
	cluster.RegisterServiceInConsul(serviceName, servicePort, servicePort)

	// 6. INICIA O SERVIDOR
	// Esta é a última chamada, pois é bloqueante.
	listenAddress := fmt.Sprintf(":%d", servicePort)
	log.Printf("[Main] Shop service HTTP server starting on %s. Handing over to HTTP listener.", listenAddress)

	if err := http.ListenAndServe(listenAddress, nil); err != nil {
		log.Fatalf("Fatal: Shop service HTTP server failed to start: %v", err)
	}
}