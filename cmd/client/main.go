// jokenpo/cmd/client/main.go
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"jokenpo/internal/network"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// --- Variáveis Globais para o Ping ---
var (
	pingStartTime time.Time
	pingMutex     sync.Mutex
)

// --- Máquina de Estados do Cliente (Inalterada) ---
const (
	StateMainMenu = "MainMenu"
	StateInQueue  = "InQueue"
	StateInMatch  = "InMatch"
)

var clientState = StateMainMenu

// --- Ponto de Entrada ---
func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	serverHost := os.Getenv("SERVER_ADDRESS")
	if serverHost == "" {
		serverHost = "localhost:80" // Conecta na porta do Traefik
	}

	u := url.URL{Scheme: "ws", Host: serverHost, Path: "/ws"}
	log.Printf("Tentando conectar ao servidor WebSocket em %s", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatalf("Não foi possível conectar: %v", err)
	}
	defer conn.Close()
	log.Println("Conexão WebSocket bem-sucedida!")

	// Canal para receber o resultado do ping do PongHandler
	pingResultChan := make(chan time.Duration)

	// --- CONFIGURANDO O PONG HANDLER ---
	// Esta função é chamada pela biblioteca quando um PONG é recebido.
	conn.SetPongHandler(func(appData string) error {
		pingMutex.Lock()
		defer pingMutex.Unlock()
		if !pingStartTime.IsZero() {
			latency := time.Since(pingStartTime)
			pingResultChan <- latency   // Envia o resultado para a função doPing
			pingStartTime = time.Time{} // Reseta o cronômetro
		}
		return nil
	})

	done := make(chan struct{})
	go readLoop(conn, done)

	// Inicia uma goroutine para lidar com o input do usuário
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			userInput := scanner.Text()
			handleUserInput(conn, scanner, userInput, pingResultChan)
		}
	}()

	// Espera por interrupção (Ctrl+C) ou desconexão
	select {
	case <-done:
		log.Println("Desconectado do servidor.")
	case <-interrupt:
		log.Println("Interrupção recebida, fechando conexão.")
		// Envia uma mensagem de fechamento limpa para o servidor
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	}
}

// --- Goroutine de Leitura ---
func readLoop(conn *websocket.Conn, done chan struct{}) {
	defer close(done)
	for {
		var msg network.Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			if !websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("\nConexão fechada normalmente.")
			} else {
				log.Printf("\nErro de leitura: %v", err)
			}
			break
		}

		if msg.Type == "RESPONSE_SUCCESS" {
			var payload struct{ State string `json:"state"` }
			json.Unmarshal(msg.Payload, &payload)
			if payload.State != "" {
				updateClientState(payload.State)
			}
		}

		printServerMessage(&msg)

		if msg.Type == "PROMPT_INPUT" {
			printPrompt()
		}
	}
}

// --- Lógica de Input ---

// handleUserInput centraliza todo o tratamento de entrada do usuário.
func handleUserInput(conn *websocket.Conn, scanner *bufio.Scanner, userInput string, pingResultChan chan time.Duration) {
	if clientState == StateMainMenu && userInput == "9" {
		// --- LÓGICA DO PING ---
		fmt.Println("\nEnviando ping...")

		pingMutex.Lock()
		pingStartTime = time.Now()
		pingMutex.Unlock()

		// Envia a mensagem de controle de PING
		err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(time.Second*5))
		if err != nil {
			log.Println("Erro ao enviar ping:", err)
			pingMutex.Lock()
			pingStartTime = time.Time{} // Reseta em caso de falha
			pingMutex.Unlock()
			return
		}

		// Espera pela resposta no canal OU por um timeout
		select {
		case latency := <-pingResultChan:
			// Usamos .Nanoseconds() para obter o valor bruto em nanosegundos
			fmt.Printf("[INFO: Pong recebido! Latência: %d ns (%v)]\n", latency.Nanoseconds(), latency)
		case <-time.After(3 * time.Second):
			fmt.Println("[ERRO: Timeout do ping. Nenhuma resposta do servidor.]")
		}
		printPrompt() // Mostra o menu novamente

	} else {
		// Lógica normal para outras entradas
		switch clientState {
		case StateMainMenu:
			handleMainMenuInput(conn, scanner, userInput)
		case StateInQueue:
			handleInQueueInput(conn, userInput)
		case StateInMatch:
			handleInMatchInput(conn, userInput)
		}
	}
}

func updateClientState(newState string) {
	switch newState {
	case "lobby":
		clientState = StateMainMenu
	case "in-queue":
		clientState = StateInQueue
	case "in-match":
		clientState = StateInMatch
	default:
		log.Printf("Alerta: Servidor enviou estado desconhecido ('%s').\n", newState)
		clientState = StateMainMenu
	}
}

func handleMainMenuInput(conn *websocket.Conn, scanner *bufio.Scanner, choice string) {
	var msg network.Message
	shouldSend := true
	switch choice {
	case "1":
		msg.Type = "FIND_MATCH"
	case "2":
		msg.Type = "PURCHASE_PACKAGE"
	case "3":
		amount, err := promptForInt(scanner, "Digite a quantidade: ")
		if err != nil {
			fmt.Println(err)
			shouldSend = false
		} else {
			payload, _ := json.Marshal(map[string]int{"amount": amount})
			msg = network.Message{Type: "PURCHASE_MULTI_PACKAGE", Payload: payload}
		}
	case "4":
		msg.Type = "VIEW_COLLECTION"
	case "5":
		msg.Type = "VIEW_DECK"
	case "6":
		key := promptForString(scanner, "Digite a chave da carta (ex: rock:5:red): ")
		payload, _ := json.Marshal(map[string]string{"key": key})
		msg = network.Message{Type: "ADD_CARD_TO_DECK", Payload: payload}
	case "7":
		index, err := promptForInt(scanner, "Digite o índice da carta a remover: ")
		if err != nil {
			fmt.Println(err)
			shouldSend = false
		} else {
			payload, _ := json.Marshal(map[string]int{"index": index})
			msg = network.Message{Type: "REMOVE_CARD_FROM_DECK", Payload: payload}
		}
	case "8":
		index, err := promptForInt(scanner, "Digite o índice da carta a substituir: ")
		if err != nil {
			fmt.Println(err)
			shouldSend = false
			break
		}
		key := promptForString(scanner, "Digite a chave da nova carta: ")
		payload, _ := json.Marshal(map[string]interface{}{"index": index, "key": key})
		msg = network.Message{Type: "REPLACE_CARD_TO_DECK", Payload: payload}
	case "9": // A lógica de ping já foi tratada, então não fazemos nada aqui.
		shouldSend = false
	default:
		fmt.Println("Opção inválida.")
		shouldSend = false
	}

	if shouldSend {
		if err := conn.WriteJSON(msg); err != nil {
			log.Printf("Erro ao enviar mensagem: %v", err)
		}
	} else if choice != "9" {
		printPrompt()
	}
}

func handleInQueueInput(conn *websocket.Conn, choice string) {
	if choice == "0" {
		msg := network.Message{Type: "LEAVE_QUEUE"}
		if err := conn.WriteJSON(msg); err != nil {
			log.Printf("Erro ao enviar mensagem: %v", err)
		}
	} else {
		fmt.Println("Opção inválida.")
		printPrompt()
	}
}

func handleInMatchInput(conn *websocket.Conn, choice string) {
	index, err := strconv.Atoi(choice)
	if err != nil {
		fmt.Println("Entrada inválida. Por favor, digite um número.")
		printPrompt()
		return
	}
	payload, _ := json.Marshal(map[string]int{"cardIndex": index})
	msg := network.Message{Type: "PLAY_CARD", Payload: payload}
	if err := conn.WriteJSON(msg); err != nil {
		log.Printf("Erro ao enviar mensagem: %v", err)
	}
}

// --- Funções de Utilidade ---
func printServerMessage(msg *network.Message) {
	if msg.Type == "PROMPT_INPUT" {
		return
	}
	var successPayload struct {
		Message string `json:"message"`
		Data    any    `json:"data"`
	}
	var errorPayload struct{ Error string `json:"error"` }

	if msg.Type == "RESPONSE_SUCCESS" && json.Unmarshal(msg.Payload, &successPayload) == nil {
		fmt.Printf("\n%s\n", successPayload.Message)

		if successPayload.Data != nil {
			// --- INÍCIO DA CORREÇÃO ---
			// Verificamos se o 'Data' é uma string.
			if strData, ok := successPayload.Data.(string); ok {
				// Se for uma string, imprimimos diretamente.
				// O Go interpretará os \n como quebras de linha.
				fmt.Println(strData)
			} else {
				// Se for qualquer outra coisa (mapa, slice, etc.), formatamos como JSON.
				prettyJSON, err := json.MarshalIndent(successPayload.Data, "", "  ")
				if err == nil {
					fmt.Println(string(prettyJSON))
				} else {
					// Fallback caso o MarshalIndent falhe
					fmt.Printf("%v\n", successPayload.Data)
				}
			}
			// --- FIM DA CORREÇÃO ---
		}
	} else if msg.Type == "RESPONSE_ERROR" && json.Unmarshal(msg.Payload, &errorPayload) == nil {
		fmt.Printf("\nErro: %s\n", errorPayload.Error)
	} else {
		// Mensagens genéricas que não se encaixam no padrão
		fmt.Printf("\nInfo (%s): %s\n", msg.Type, string(msg.Payload))
	}
}

func printPrompt() {
	var prompt string
	time.Sleep(1000 * time.Millisecond)
	switch clientState {
	case StateMainMenu:
		prompt = `
--- Jokenpo Card Game (Lobby) ---
1. Buscar Partida
2. Comprar Pacote
3. Comprar Múltiplos Pacotes
4. Ver Coleção
5. Ver Deck
6. Adicionar Carta ao Deck
7. Remover Carta do Deck
8. Substituir Carta no Deck
9. Medir Ping (WebSocket)
---------------------------------

(Lobby) Digite uma opção: `
	case StateInQueue:
		prompt = "\n(Na Fila) Digite 0 para sair: "
	case StateInMatch:
		prompt = "\n(Em Jogo) Digite o índice da carta para jogar: "
	}
	fmt.Print(prompt)
}

func promptForString(scanner *bufio.Scanner, prompt string) string {
	fmt.Print(prompt)
	scanner.Scan()
	return scanner.Text()
}

func promptForInt(scanner *bufio.Scanner, prompt string) (int, error) {
	fmt.Print(prompt)
	scanner.Scan()
	input := scanner.Text()
	num, err := strconv.Atoi(input)
	if err != nil {
		return 0, fmt.Errorf("entrada inválida. Por favor, digite um número")
	}
	return num, nil
}