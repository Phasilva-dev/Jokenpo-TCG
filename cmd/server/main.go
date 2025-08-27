package main

import (
	"encoding/json"
	"log"
	"sync"

	"jokenpo/internal/network" // Verifique se este é o path correto
)

// GameHandler agora mantém o estado do número de clientes de forma segura.
type GameHandler struct {
	mu          sync.Mutex
	clientCount int
}

func (h *GameHandler) OnConnect(c *network.Client) {
	h.mu.Lock()
	h.clientCount++
	h.mu.Unlock()
	log.Printf("[GameHandler] Novo cliente conectado: %s. Total: %d\n", c.Conn().RemoteAddr(), h.clientCount)
}

func (h *GameHandler) OnDisconnect(c *network.Client) {
	h.mu.Lock()
	h.clientCount--
	h.mu.Unlock()
	log.Printf("[GameHandler] Cliente desconectado: %s. Total: %d\n", c.Conn().RemoteAddr(), h.clientCount)
}

func (h *GameHandler) OnMessage(c *network.Client, msg network.Message) {
	log.Printf("[GameHandler] Mensagem de %s: Tipo=%s\n", c.Conn().RemoteAddr(), msg.Type)

	switch msg.Type {
	case "TEST":
		// Simplesmente ecoa a mensagem de teste de volta para o remetente
		c.Send() <- msg

	case "GET_CLIENT_COUNT":
		h.mu.Lock()
		count := h.clientCount
		h.mu.Unlock()

		// Cria um payload de resposta com a contagem atual
		payload := map[string]int{"count": count}
		payloadBytes, _ := json.Marshal(payload)

		// Cria e envia a mensagem de resposta
		response := network.Message{
			Type:    "CLIENT_COUNT_RESPONSE",
			Payload: payloadBytes,
		}
		c.Send() <- response
	}
}

func main() {
	gameHandler := &GameHandler{}
	server := network.NewServer(gameHandler)
	err := server.Listen(":8080")
	if err != nil {
		log.Fatalf("Não foi possível iniciar o servidor: %v", err)
	}
}