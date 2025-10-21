package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"jokenpo/internal/network"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	pingStartTime time.Time
	pingMutex     sync.Mutex
)

// --- MUDANÇA: Adicionado novo estado ---
const (
	StateMainMenu     = "MainMenu"
	StateInQueue      = "InQueue"
	StateInTradeQueue = "InTradeQueue" // Novo estado
	StateInMatch      = "InMatch"
)

var clientState = StateMainMenu

func main() {
	// ... (função main sem mudanças) ...
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// --- INÍCIO DA LÓGICA DE FAILOVER ---

	// 1. Define a lista de endereços dos Load Balancers
	lbAddresses := []string{
		"localhost:9080",
		"localhost:9081",
		"localhost:9082",
	}

	// Permite que a lista seja sobrescrita por uma variável de ambiente (para flexibilidade)
	// Ex: LB_ADDRESSES="192.168.1.10:80,192.168.1.11:80"
	if addrsEnv := os.Getenv("LB_ADDRESSES"); addrsEnv != "" {
		lbAddresses = strings.Split(addrsEnv, ",")
	}

	var conn *websocket.Conn
	var err error

	// 2. Tenta conectar em cada endereço da lista até ter sucesso
	for _, addr := range lbAddresses {
		u := url.URL{Scheme: "ws", Host: strings.TrimSpace(addr), Path: "/ws"}
		log.Printf("Tentando conectar ao Load Balancer em %s", u.String())

		// Tenta a conexão. O 'nil' para o header é importante.
		var resp *http.Response // Captura a resposta para depuração
		conn, resp, err = websocket.DefaultDialer.Dial(u.String(), nil)
		if err == nil {
			// Conexão bem-sucedida!
			log.Println("Conexão WebSocket bem-sucedida!")
			break // Sai do loop
		}

		// Se a conexão falhou, loga o motivo e tenta o próximo
		log.Printf("AVISO: Falha ao conectar a %s: %v", addr, err)
		if resp != nil {
			log.Printf("AVISO: Status da resposta recebida: %s", resp.Status)
		}
	}

	// 3. Se após o loop a conexão ainda for nula, todos os LBs falharam.
	if conn == nil {
		log.Fatalf("Não foi possível conectar a nenhum dos Load Balancers disponíveis. Encerrando.")
	}
	defer conn.Close()

	// --- FIM DA LÓGICA DE FAILOVER ---

	pingResultChan := make(chan time.Duration)
	conn.SetPongHandler(func(appData string) error {
		pingMutex.Lock()
		defer pingMutex.Unlock()
		if !pingStartTime.IsZero() {
			latency := time.Since(pingStartTime)
			pingResultChan <- latency
			pingStartTime = time.Time{}
		}
		return nil
	})

	done := make(chan struct{})
	go readLoop(conn, done)

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			userInput := scanner.Text()
			handleUserInput(conn, scanner, userInput, pingResultChan)
		}
	}()

	select {
	case <-done:
		log.Println("Desconectado do servidor.")
	case <-interrupt:
		log.Println("Interrupção recebida, fechando conexão.")
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	}
}

func readLoop(conn *websocket.Conn, done chan struct{}) {
	// ... (função readLoop sem mudanças) ...
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

func handleUserInput(conn *websocket.Conn, scanner *bufio.Scanner, userInput string, pingResultChan chan time.Duration) {
	// --- MUDANÇA: Adicionado case para o novo estado ---
	if (clientState == StateInQueue || clientState == StateInTradeQueue) && userInput == "0" {
		// Comando genérico para sair de qualquer fila
		msg := network.Message{Type: "LEAVE_QUEUE"}
		if err := conn.WriteJSON(msg); err != nil {
			log.Printf("Erro ao enviar mensagem: %v", err)
		}
	} else if clientState == StateMainMenu && userInput == "9" {
		// ... (lógica do ping, sem mudanças)
		fmt.Println("\nEnviando ping...")

		pingMutex.Lock()
		pingStartTime = time.Now()
		pingMutex.Unlock()

		err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(time.Second*5))
		if err != nil {
			log.Println("Erro ao enviar ping:", err)
			pingMutex.Lock()
			pingStartTime = time.Time{}
			pingMutex.Unlock()
			return
		}

		select {
		case latency := <-pingResultChan:
			fmt.Printf("[INFO: Pong recebido! Latência: %d ns (%v)]\n", latency.Nanoseconds(), latency)
		case <-time.After(3 * time.Second):
			fmt.Println("[ERRO: Timeout do ping. Nenhuma resposta do servidor.]")
		}
		printPrompt()
	} else {
		// Roteia para o handler correto com base no estado.
		switch clientState {
		case StateMainMenu:
			handleMainMenuInput(conn, scanner, userInput)
		case StateInQueue:
			handleInQueueInput(conn, userInput)
		case StateInTradeQueue:
			handleInTradeQueueInput(conn, userInput) // Novo handler
		case StateInMatch:
			handleInMatchInput(conn, userInput)
		}
	}
}

// --- MUDANÇA: Adicionado case para o novo estado ---
func updateClientState(newState string) {
	switch newState {
	case "lobby":
		clientState = StateMainMenu
	case "in-match-queue": // Nome do estado atualizado
		clientState = StateInQueue
	case "in-trade-queue": // Novo estado
		clientState = StateInTradeQueue
	case "in-match":
		clientState = StateInMatch
	default:
		log.Printf("Alerta: Servidor enviou estado desconhecido ('%s').\n", newState)
		clientState = StateMainMenu
	}
}

// --- MUDANÇA: Lógica de "Trocar Carta" adicionada ---
func handleMainMenuInput(conn *websocket.Conn, scanner *bufio.Scanner, choice string) {
	var msg network.Message
	shouldSend := true
	switch choice {
	case "1":
		msg.Type = "FIND_MATCH"
	case "2":
		// Solicita ao jogador a carta que ele quer oferecer.
		cardKey := promptForString(scanner, "Digite a chave da carta que você quer trocar (ex: rock:5:red): ")
		if cardKey == "" {
			fmt.Println("A chave da carta não pode ser vazia.")
			shouldSend = false
		} else {
			// Cria o payload com o campo "cardKey" que o servidor espera.
			payload, _ := json.Marshal(map[string]string{"cardKey": cardKey})
			msg = network.Message{Type: "TRADE_CARD", Payload: payload}
		}
	case "3":
		amount, err := promptForInt(scanner, "Digite a quantidade: ")
		if err != nil {
			fmt.Println(err)
			shouldSend = false
		} else {
			payload, _ := json.Marshal(map[string]int{"quantity": amount})
			msg = network.Message{Type: "PURCHASE_PACKAGE", Payload: payload}
		}
	// ... (resto dos cases sem mudanças)
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
	case "9":
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

// Unificado: a única opção é sair.
func handleInQueueInput(conn *websocket.Conn, choice string) {
	if choice != "0" {
		fmt.Println("Opção inválida.")
		printPrompt()
	}
	// A lógica de envio foi movida para handleUserInput
}

// --- NOVO HANDLER ---
// A única opção na fila de troca também é sair.
func handleInTradeQueueInput(conn *websocket.Conn, choice string) {
	if choice != "0" {
		fmt.Println("Opção inválida.")
		printPrompt()
	}
	// A lógica de envio foi movida para handleUserInput
}


func handleInMatchInput(conn *websocket.Conn, choice string) {
	// ... (função sem mudanças) ...
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

func printServerMessage(msg *network.Message) {
	// ... (função sem mudanças) ...
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
			if strData, ok := successPayload.Data.(string); ok {
				fmt.Println(strData)
			} else {
				prettyJSON, err := json.MarshalIndent(successPayload.Data, "", "  ")
				if err == nil {
					fmt.Println(string(prettyJSON))
				} else {
					fmt.Printf("%v\n", successPayload.Data)
				}
			}
		}
	} else if msg.Type == "RESPONSE_ERROR" && json.Unmarshal(msg.Payload, &errorPayload) == nil {
		fmt.Printf("\nErro: %s\n", errorPayload.Error)
	} else {
		fmt.Printf("\nInfo (%s): %s\n", msg.Type, string(msg.Payload))
	}
}

// --- MUDANÇA: Adicionado menu para o novo estado ---
func printPrompt() {
	var prompt string
	time.Sleep(100 * time.Millisecond) // Reduzido para uma melhor experiência
	switch clientState {
	case StateMainMenu:
		prompt = `
--- Jokenpo Card Game (Lobby) ---
1. Buscar Partida
2. Trocar Carta (Wonder Trade)
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
		prompt = "\n(Na Fila de Partida) Digite 0 para sair: "
	case StateInTradeQueue: // Novo menu
		prompt = "\n(Na Fila de Troca) Digite 0 para sair: "
	case StateInMatch:
		prompt = "\n(Em Jogo) Digite o índice da carta para jogar: "
	}
	fmt.Print(prompt)
}

func promptForString(scanner *bufio.Scanner, prompt string) string {
	// ... (função sem mudanças) ...
	fmt.Print(prompt)
	scanner.Scan()
	return scanner.Text()
}
func promptForInt(scanner *bufio.Scanner, prompt string) (int, error) {
	// ... (função sem mudanças) ...
	fmt.Print(prompt)
	scanner.Scan()
	input := scanner.Text()
	num, err := strconv.Atoi(input)
	if err != nil {
		return 0, fmt.Errorf("entrada inválida. Por favor, digite um número")
	}
	return num, nil
}