// jokenpo/cmd/client/main.go
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"jokenpo/internal/network"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

// --- Máquina de Estados do Cliente ---
const (
	StateMainMenu = "MainMenu"
	StateInQueue  = "InQueue"
	StateInMatch  = "InMatch"
)

var clientState = StateMainMenu

// --- Ponto de Entrada ---
func main() {
	address := "localhost:8080"
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatalf("Não foi possível conectar ao servidor: %v", err)
	}
	defer conn.Close()

	// A goroutine de leitura é responsável por receber mensagens e ATUALIZAR o estado.
	go readLoop(conn)

	// A goroutine principal tem um único loop para ler o input do usuário.
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		userInput := scanner.Text()

		// Após receber o input, ele o despacha para o handler correto
		// com base no estado ATUAL, que foi atualizado pelo readLoop.
		switch clientState {
		case StateMainMenu:
			handleMainMenuInput(conn, scanner, userInput)
		case StateInQueue:
			handleInQueueInput(conn, userInput)//, scanner
		case StateInMatch:
			handleInMatchInput(conn, userInput)//, scanner
		}
	}
}

// --- Goroutine de Leitura ---

func readLoop(conn net.Conn) {
	for {
		msg, err := network.ReadMessage(conn)
		if err != nil {
			log.Println("\nConexão com o servidor perdida.")
			os.Exit(0)
		}

		// *** LÓGICA CRÍTICA DE MUDANÇA DE ESTADO ***
		if msg.Type == "RESPONSE_SUCCESS" {
			var payload struct{ Message string `json:"message"` }
			json.Unmarshal(msg.Payload, &payload)
			
			if strings.Contains(payload.Message, "matchmaking queue") {
				clientState = StateInQueue
			} else if strings.Contains(payload.Message, "Match found!") || strings.Contains(payload.Message, "The round has started!") {
				clientState = StateInMatch
			} else if strings.Contains(payload.Message, "You have returned to the lobby") {
				clientState = StateMainMenu
			}
		}

		// Apenas imprime a mensagem do servidor e depois o prompt correto.
		printServerMessage(msg)
		printPrompt()
	}
}

// --- Handlers de Input (Recebem o input, não o leem) ---

func handleMainMenuInput(conn net.Conn, scanner *bufio.Scanner, choice string) {
	var msg network.Message
	shouldSend := true

	switch choice {
	case "1": msg.Type = "FIND_MATCH"
	case "2": msg.Type = "PURCHASE_PACKAGE"
	case "3":
		amount, err := promptForInt(scanner, "Digite a quantidade: ")
		if err != nil { fmt.Println(err); printPrompt(); shouldSend = false; return }
		payload, _ := json.Marshal(map[string]int{"amount": amount})
		msg = network.Message{Type: "PURCHASE_MULTI_PACKAGE", Payload: payload}
	case "4": msg.Type = "VIEW_COLLECTION"
	case "5": msg.Type = "VIEW_DECK"
	case "6":
		key := promptForString(scanner, "Digite a chave da carta (ex: rock:5:red): ")
		payload, _ := json.Marshal(map[string]string{"key": key})
		msg = network.Message{Type: "ADD_CARD_TO_DECK", Payload: payload}
	case "7":
		index, err := promptForInt(scanner, "Digite o índice da carta a remover: ")
		if err != nil { fmt.Println(err); printPrompt(); shouldSend = false; return }
		payload, _ := json.Marshal(map[string]int{"index": index})
		msg = network.Message{Type: "REMOVE_CARD_FROM_DECK", Payload: payload}
	case "8":
		index, err := promptForInt(scanner, "Digite o índice da carta a substituir: ")
		if err != nil { fmt.Println(err); printPrompt(); shouldSend = false; return }
		key := promptForString(scanner, "Digite a chave da nova carta: ")
		payload, _ := json.Marshal(map[string]interface{}{"index": index, "key": key})
		msg = network.Message{Type: "REPLACE_CARD_TO_DECK", Payload: payload}
	default:
		fmt.Println("Opção inválida.")
		printPrompt()
		shouldSend = false
	}

	if shouldSend {
		if err := network.WriteMessage(conn, msg); err != nil {
			log.Printf("Erro ao enviar mensagem: %v", err)
		}
	}
}

func handleInQueueInput(conn net.Conn, choice string) {
	if choice == "0" {
		msg := network.Message{Type: "LEAVE_QUEUE"}
		if err := network.WriteMessage(conn, msg); err != nil {
			log.Printf("Erro ao enviar mensagem: %v", err)
		}
	} else {
		fmt.Println("Opção inválida.")
		printPrompt()
	}
}

func handleInMatchInput(conn net.Conn , choice string) {//scanner *bufio.Scanner
	index, err := strconv.Atoi(choice)
	if err != nil {
		fmt.Println("Entrada inválida. Por favor, digite um número.")
		printPrompt()
		return
	}

	payload, _ := json.Marshal(map[string]int{"cardIndex": index})
	msg := network.Message{Type: "PLAY_CARD", Payload: payload}
	if err := network.WriteMessage(conn, msg); err != nil {
		log.Printf("Erro ao enviar mensagem: %v", err)
	}
}

// --- Funções de Utilidade ---

// printServerMessage é a única função que foi alterada nesta versão.
func printServerMessage(msg *network.Message) {
	var successPayload struct {
		Message string `json:"message"`
		Data    any    `json:"data"`
	}
	var errorPayload struct {
		Error string `json:"error"`
	}

	if msg.Type == "RESPONSE_SUCCESS" && json.Unmarshal(msg.Payload, &successPayload) == nil {
		fmt.Printf("\n%s\n", successPayload.Message) // Saída limpa
		if successPayload.Data != nil {
			fmt.Printf("%v\n", successPayload.Data) // Saída limpa
		}
	} else if msg.Type == "RESPONSE_ERROR" && json.Unmarshal(msg.Payload, &errorPayload) == nil {
		fmt.Printf("\nErro: %s\n", errorPayload.Error) // Saída limpa
	} else {
		// Para mensagens desconhecidas, ainda é útil ver os detalhes para depuração.
		fmt.Printf("\nInfo (%s): %s\n", msg.Type, string(msg.Payload))
	}
}


// printPrompt é chamado após qualquer mensagem do servidor para guiar o usuário.
func printPrompt() {
	switch clientState {
	case StateMainMenu:
		fmt.Print("\n(Lobby) Digite uma opção (1-8): ")
	case StateInQueue:
		fmt.Print("\n(Na Fila) Digite 0 para sair: ")
	case StateInMatch:
		fmt.Print("\n(Em Jogo) Digite o índice da carta para jogar: ")
	}
}


// As funções para pedir input adicional não mudam.
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