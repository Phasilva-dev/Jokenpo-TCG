// cmd/server/main.go
package main

import (
	"fmt"
	"log"
	"jokenpo/internal/network" // Verifique se o nome do módulo está correto!
)

// GameHandler é a nossa implementação da interface network.EventHandler.
// É aqui que TODA a lógica do seu jogo vai viver.
type GameHandler struct {
	// Futuramente: mapas de jogadores, salas de jogo, etc.
}

// OnConnect é chamado pelo pacote de rede quando um novo cliente se conecta.
func (h *GameHandler) OnConnect(c *network.Client) {
	fmt.Printf("[GameHandler] Novo cliente conectado: %s\n", c.Conn().RemoteAddr())
}

// OnDisconnect é chamado quando um cliente se desconecta.
func (h *GameHandler) OnDisconnect(c *network.Client) {
	fmt.Printf("[GameHandler] Cliente desconectado: %s\n", c.Conn().RemoteAddr())
}

// OnMessage é chamado para cada mensagem recebida de um cliente.
func (h *GameHandler) OnMessage(c *network.Client, msg network.Message) {
	fmt.Printf("[GameHandler] Mensagem de %s: Tipo=%s, Payload=%s\n", c.Conn().RemoteAddr(), msg.Type, string(msg.Payload))

	// Exemplo de como você pode responder ou retransmitir uma mensagem:
	// Vamos simplesmente ecoar a mensagem de volta para o cliente que a enviou.
	
	// ATENÇÃO: Nunca escreva diretamente na conexão aqui.
	// Use o canal 'send' do cliente. É seguro para concorrência.
	c.Send() <- msg
}

func main() {
	// 1. Crie uma instância da sua lógica de jogo.
	gameHandler := &GameHandler{}

	// 2. Crie um novo servidor, injetando sua lógica de jogo nele.
	server := network.NewServer(gameHandler)

	// 3. Inicie o servidor.
	err := server.Listen(":8080")
	if err != nil {
		log.Fatalf("Não foi possível iniciar o servidor: %v", err)
	}
}