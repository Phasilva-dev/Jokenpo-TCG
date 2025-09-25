// jokenpo/cmd/bot/main.go
package main

import (
	"encoding/json"
	"jokenpo/internal/network"
	"log"
	"math/rand"
	"net"
	"os"
	"time"
)

const serverAddress = "server:8080"
const pingServerAddress = "server:8081"

func main() {
	// Seeder para aleatoriedade
	rand.Seed(time.Now().UnixNano())

	// Lê a "personalidade" do bot da variável de ambiente.
	role := os.Getenv("BOT_ROLE")
	if role == "" {
		log.Fatal("FATAL: BOT_ROLE environment variable not set.")
	}

	// Conecta ao servidor TCP.
	conn, err := net.DialTimeout("tcp", serverAddress, 10*time.Second)
	if err != nil {
		log.Printf("FAIL (%s): Could not connect: %v\n", role, err)
		return
	}
	defer conn.Close()

	// Realiza o login. Se falhar, o bot encerra.
	if !waitForPrompt(conn, "Login") {
		return
	}
	log.Printf("SUCCESS (%s): Login complete. Starting main loop.", role)

	// Executa a rotina principal com base na personalidade.
	switch role {
	case "PACK_OPENER":
		runPackOpener(conn)
	case "PINGER":
		runPinger(conn)
	case "MATCHMAKER":
		runMatchmaker(conn)
	default:
		log.Fatalf("FATAL: Unknown BOT_ROLE '%s'", role)
	}
}

// --- Rotinas de Personalidade ---

// runPackOpener simula um jogador que compra pacotes repetidamente.
func runPackOpener(conn net.Conn) {
	for {
		// IMPORTANTE: Enviar 1000 de uma vez criaria uma mensagem enorme e seria rejeitado.
		// Em vez disso, simulamos o comportamento comprando 10 pacotes por vez, em um loop.
		// Isso gera um tráfego de rede mais realista e constante.
		purchaseAmount := 10
		payload, _ := json.Marshal(map[string]int{"amount": purchaseAmount})
		msg := network.Message{Type: "PURCHASE_MULTI_PACKAGE", Payload: payload}

		if err := network.WriteMessage(conn, msg); err != nil {
			log.Printf("FAIL (PACK_OPENER): Could not send purchase command: %v\n", err)
			return
		}
		log.Printf("SUCCESS (PACK_OPENER): Sent request to buy %d packs.", purchaseAmount)

		// Espera a confirmação do servidor.
		if !waitForPrompt(conn, "Purchase") {
			return
		}

		time.Sleep(time.Duration(2+rand.Intn(3)) * time.Second) // Pensa por 2-4 segundos
	}
}

// runPinger simula um jogador que mede a latência repetidamente.
func runPinger(conn net.Conn) {
	// Lança uma goroutine para descartar mensagens TCP e manter a conexão viva.
	go func() {
		for {
			_, err := network.ReadMessage(conn)
			if err != nil {
				return
			}
		}
	}()

	for {
		// A lógica de ping UDP que criamos anteriormente.
		doPing(pingServerAddress)
		time.Sleep(5 * time.Second) // Mede o ping a cada 5 segundos
	}
}

// runMatchmaker simula um jogador que entra na fila e espera.
func runMatchmaker(conn net.Conn) {
	msg := network.Message{Type: "FIND_MATCH"}
	if err := network.WriteMessage(conn, msg); err != nil {
		log.Printf("FAIL (MATCHMAKER): Could not send find_match command: %v\n", err)
		return
	}

	// O bot agora apenas fica lendo mensagens do servidor para sempre,
	// simulando um jogador que está na fila ou em uma partida ociosa.
	for {
		if _, err := network.ReadMessage(conn); err != nil {
			log.Printf("FAIL (MATCHMAKER): Connection lost while in queue/match: %v\n", err)
			return
		}
	}
}

// --- Funções de Utilidade ---

func waitForPrompt(conn net.Conn, context string) bool {
	for {
		conn.SetReadDeadline(time.Now().Add(300 * time.Second))
		msg, err := network.ReadMessage(conn)
		if err != nil {
			log.Printf("FAIL (%s): Did not receive prompt in time: %v\n", context, err)
			return false
		}
		if msg.Type == "PROMPT_INPUT" {
			return true
		}
	}
}

func doPing(serverAddress string) {
	serverAddr, err := net.ResolveUDPAddr("udp", serverAddress)
	if err != nil { return }

	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil { return }
	defer conn.Close()

	startTime := time.Now()
	pingPacket := network.EncodePingPacket(network.PING_PACKET_TYPE, startTime.UnixNano())
	conn.Write(pingPacket)

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	buffer := make([]byte, 9)
	n, _, err := conn.ReadFromUDP(buffer)
	if err != nil { return }

	packetType, _, err := network.DecodePingPacket(buffer[:n])
	if err != nil || packetType != network.PONG_PACKET_TYPE { return }

	latency := time.Since(startTime)
	log.Printf("SUCCESS (PINGER): Ping successful, latency: %v", latency)
}