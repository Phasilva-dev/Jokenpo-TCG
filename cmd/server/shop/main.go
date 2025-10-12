package main

import (
	"fmt"
	"jokenpo/internal/services/cluster"             // Para RegisterServiceInConsul e HealthAggregator
	"jokenpo/internal/services/shop"        // Para o ShopService (ator)
	"jokenpo/internal/api"    // Para o CreateShopHandler
	"log"
	"net/http"
)

// --- Configuração do Serviço ---
const (
	// O nome que o Broker usará para encontrar este serviço no Consul.
	serviceName = "jokenpo-shop"
	// A porta onde este serviço irá escutar por requisições HTTP.
	servicePort = 8081
)

func main() {
	// 1. INICIAR A LÓGICA DE NEGÓCIO (O ATOR)
	// Esta é a instância principal do nosso serviço, que gerencia o estado de forma segura.
	shopService := shop.NewShopService()
	log.Println("Shop service actor started successfully.")

	// 2. CONFIGURAR O HEALTH CHECK ESPECÍFICO
	// Em vez de um handler genérico, usamos o HealthAggregator para criar uma verificação robusta.
	health := cluster.NewHealthAggregator()

	// Adicionamos uma verificação que chama o método CheckHealth() do nosso ator.
	// Isso confirma não apenas que o servidor HTTP está online, mas que a goroutine
	// principal do serviço está viva e respondendo.
	health.AddCheck("actor_goroutine", shopService.CheckHealth)

	// 3. REGISTRAR OS HANDLERS HTTP
	// Criamos o handler principal que lida com as compras, injetando o ator.
	shopHandler := api.CreateShopHandler(shopService)

	// Associamos as rotas às suas respectivas funções de handler.
	http.HandleFunc("/Purchase", shopHandler)      // Rota principal da API
	http.HandleFunc("/health", health.Handler())   // Rota de health check para o Consul, agora usando o agregador.

	// 4. REGISTRAR O SERVIÇO NO CONSUL
	// Avisamos ao Consul que nosso serviço está prestes a ficar online.
	log.Println("Registering service with Consul...")
	cluster.RegisterServiceInConsul(serviceName, servicePort, servicePort)

	// 5. INICIAR O SERVIDOR HTTP
	// Esta é a última chamada, pois http.ListenAndServe é uma operação bloqueante.
	listenAddress := fmt.Sprintf(":%d", servicePort)
	log.Printf("Shop service HTTP server starting on %s", listenAddress)

	if err := http.ListenAndServe(listenAddress, nil); err != nil {
		log.Fatalf("Fatal: Shop service HTTP server failed to start: %v", err)
	}
}