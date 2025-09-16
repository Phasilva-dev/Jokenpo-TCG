// jokenpo/cmd/bot/main.go
package main

import (
	"jokenpo/internal/network"
	"log"
	"math/rand"
	"net"
	"time"
)

const serverAddress = "server:8080"

func main() {
	// Seeder para o gerador de números aleatórios
	rand.Seed(time.Now().UnixNano())

	conn, err := net.DialTimeout("tcp", serverAddress, 5*time.Second)
	if err != nil {
		log.Println("Connection FAIL: could not connect to server:", err)
		return
	}
	defer conn.Close()

	// --- Etapa 1: Handshake de Login ---
	// Espera pelo primeiro prompt para confirmar que o login foi bem-sucedido.
	if !waitForPrompt(conn, "Login") {
		return // Falha no login, encerra o bot.
	}
	log.Println("Login SUCCESS. Starting active command loop.")

	// --- Etapa 2: Loop de Ações Contínuas ---
	// Este loop simula um jogador ativo.
	for {
		// Envia um comando para comprar um pacote.
		purchaseMsg := network.Message{Type: "PURCHASE_PACKAGE"}
		if err := network.WriteMessage(conn, purchaseMsg); err != nil {
			log.Printf("FAIL: Could not send purchase command: %v\n", err)
			return // Se não conseguimos escrever, a conexão provavelmente caiu.
		}

		// Espera pelo próximo prompt, que sinaliza que a ação de compra foi concluída.
		if !waitForPrompt(conn, "Purchase Action") {
			return // A conexão provavelmente caiu durante a espera.
		}

		// Simula um jogador pensando ou olhando suas cartas antes da próxima ação.
		// Um delay aleatório torna o teste ainda mais realista.
		thinkTime := time.Duration(2+rand.Intn(4)) * time.Second // Espera entre 2 e 5 segundos
		time.Sleep(thinkTime)
	}
}

// waitForPrompt é uma função de utilidade que lê da conexão até
// encontrar uma mensagem PROMPT_INPUT ou atingir um timeout.
// Retorna 'true' se o prompt foi recebido, 'false' se houve falha.
func waitForPrompt(conn net.Conn, context string) bool {
	for {
		// Um timeout agressivo para cada leitura. Se o servidor não responder
		// rapidamente, consideramos uma falha.
		conn.SetReadDeadline(time.Now().Add(10 * time.Second))

		msg, err := network.ReadMessage(conn)
		if err != nil {
			log.Printf("FAIL during '%s': did not receive server response in time: %v\n", context, err)
			return false
		}

		// Encontramos o sinal do servidor para agir. A missão foi cumprida.
		if msg.Type == "PROMPT_INPUT" {
			return true
		}
		// Se for outra mensagem (RESPONSE_SUCCESS), nós a ignoramos e continuamos
		// lendo até o PROMPT_INPUT chegar.
	}
}