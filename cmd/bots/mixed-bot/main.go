// jokenpo/cmd/bot/main.go
package main

import (
	"encoding/json"
	"jokenpo/internal/network"
	"log"
	"math/rand"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// O endereço agora é apenas o host:porta, o protocolo 'ws://' será adicionado.
const serverAddress = "server:8080"

// Variáveis globais para o bot PINGER
var (
	pingStartTime time.Time
	pingMutex     sync.Mutex
)

func main() {
	rand.Seed(time.Now().UnixNano())

	role := os.Getenv("BOT_ROLE")
	if role == "" {
		log.Fatal("FATAL: BOT_ROLE environment variable not set.")
	}

	// --- LÓGICA DE CONEXÃO ATUALIZADA ---
	u := url.URL{Scheme: "ws", Host: serverAddress, Path: "/ws"}
	log.Printf("INFO (%s): Connecting to WebSocket server at %s", role, u.String())

	// Usa o Dialer do WebSocket em vez de net.Dial
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Printf("FAIL (%s): Could not connect: %v\n", role, err)
		return
	}
	defer conn.Close()

	// Se o bot for um PINGER, configuramos o PongHandler para medir a latência.
	if role == "PINGER" {
		conn.SetPongHandler(func(appData string) error {
			pingMutex.Lock()
			latency := time.Since(pingStartTime)
			pingMutex.Unlock()
			log.Printf("SUCCESS (PINGER): Pong received, latency: %v", latency)
			return nil
		})
	}

	// Espera pelo prompt inicial. A função foi adaptada para websocket.Conn
	if !waitForPrompt(conn, "Login") {
		return
	}
	log.Printf("SUCCESS (%s): Login complete. Starting main loop.", role)

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

// --- Rotinas de Personalidade (Atualizadas) ---

// runPackOpener agora usa conn.WriteJSON
func runPackOpener(conn *websocket.Conn) {
	for {
		purchaseAmount := 10
		payload, _ := json.Marshal(map[string]int{"amount": purchaseAmount})
		msg := network.Message{Type: "PURCHASE_MULTI_PACKAGE", Payload: payload}

		if err := conn.WriteJSON(msg); err != nil {
			log.Printf("FAIL (PACK_OPENER): Could not send purchase command: %v\n", err)
			return
		}
		log.Printf("SUCCESS (PACK_OPENER): Sent request to buy %d packs.", purchaseAmount)

		if !waitForPrompt(conn, "Purchase") {
			return
		}
		time.Sleep(time.Duration(2+rand.Intn(3)) * time.Second)
	}
}

// runPinger foi completamente reescrito para usar o ping do WebSocket.
func runPinger(conn *websocket.Conn) {
	// Lança uma goroutine para descartar mensagens (necessário para receber pongs)
	go func() {
		for {
			var msg network.Message
			if err := conn.ReadJSON(&msg); err != nil {
				return // Encerra a goroutine se a conexão fechar
			}
		}
	}()

	// Loop principal para enviar pings
	for {
		pingMutex.Lock()
		pingStartTime = time.Now()
		pingMutex.Unlock()

		// Envia uma mensagem de controle de PING do WebSocket
		err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(time.Second*5))
		if err != nil {
			log.Printf("FAIL (PINGER): Could not send ping: %v", err)
			return
		}
		
		time.Sleep(5 * time.Second)
	}
}

// runMatchmaker agora usa conn.WriteJSON e conn.ReadJSON
func runMatchmaker(conn *websocket.Conn) {
	msg := network.Message{Type: "FIND_MATCH"}
	if err := conn.WriteJSON(msg); err != nil {
		log.Printf("FAIL (MATCHMAKER): Could not send find_match command: %v\n", err)
		return
	}

	// O bot fica lendo mensagens do servidor para sempre
	for {
		var dummyMsg network.Message // Usado apenas para descartar a mensagem
		if err := conn.ReadJSON(&dummyMsg); err != nil {
			log.Printf("FAIL (MATCHMAKER): Connection lost: %v\n", err)
			return
		}
	}
}

// --- Funções de Utilidade (Atualizadas) ---

// waitForPrompt agora usa conn.ReadJSON
func waitForPrompt(conn *websocket.Conn, context string) bool {
	for {
		// A biblioteca websocket já tem deadlines, mas podemos adicionar um se quisermos
		var msg network.Message
		if err := conn.ReadJSON(&msg); err != nil {
			log.Printf("FAIL (%s): Did not receive prompt: %v\n", context, err)
			return false
		}

		if msg.Type == "PROMPT_INPUT" {
			return true
		}
	}
}
