//START OF FILE jokenpo/cmd/server/gameroom/main.go
package main
/*
import (
	"fmt"
	"jokenpo/internal/services/cluster"
	"jokenpo/internal/services/gameroom"
	"log"
	"net/http"
	"os"
)

const (
	serviceName = "jokenpo-gameroom"
	servicePort = 8083 // Porta dedicada para este serviço
)

func main() {
	log.Println("Starting Jokenpo GameRoom Service instance...")

	// 1. Carrega a configuração (endereço do Consul)
	consulAddr := os.Getenv("CONSUL_HTTP_ADDR")
	if consulAddr == "" {
		consulAddr = "consul-1:8500" // Padrão para Docker
	}
	// O endereço que este serviço anunciará para os outros.
	advertiseAddr := os.Getenv("ADVERTISE_ADDR")
	if advertiseAddr == "" {
		hostname, _ := os.Hostname()
		advertiseAddr = hostname
	}

	// 2. Cria o RoomManager, que irá criar e supervisionar as salas.
	roomManager := gameroom.NewRoomManager()
	go roomManager.Run() // Inicia a goroutine do ator RoomManager
	log.Println("[Main] RoomManager actor created and started.")

	// 3. Configura os handlers da API HTTP
	mux := http.NewServeMux()
	mux.HandleFunc("/health", cluster.NewBasicHealthHandler())
	// Passa o endereço de anúncio para que a API possa retorná-lo.
	api.RegisterHandlers(mux, roomManager, advertiseAddr, servicePort)
	log.Println("[Main] HTTP handlers registered for /rooms and /health.")

	// 4. Registra o serviço no Consul
	log.Println("[Main] Registering service with Consul...")
	err := cluster.RegisterServiceInConsul(serviceName, servicePort, servicePort, consulAddr)
	if err != nil {
		log.Fatalf("Fatal: Falha ao registrar serviço no Consul: %v", err)
	}

	// 5. Inicia o servidor HTTP
	listenAddress := fmt.Sprintf(":%d", servicePort)
	log.Printf("[Main] GameRoom service HTTP server starting on %s.", listenAddress)
	if err := http.ListenAndServe(listenAddress, mux); err != nil {
		log.Fatalf("Fatal: Falha ao iniciar servidor HTTP: %v", err)
	}
}
*/
//END OF FILE jokenpo/cmd/server/gameroom/main.go