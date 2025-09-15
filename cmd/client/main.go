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
	StateMainMenu = "MainMenu" // O jogador está no lobby principal.
	StateInQueue  = "InQueue"  // O jogador está na fila de matchmaking.
	StateInMatch  = "InMatch"  // O jogador está em uma partida.
)

// Esta variável global controla qual menu é exibido para o usuário.
var clientState = StateMainMenu

// --- Ponto de Entrada ---
func main() {
	address := "localhost:8080"
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatalf("Não foi possível conectar ao servidor: %v", err)
	}
	defer conn.Close()

	// A goroutine de leitura ATUALIZA o estado do cliente.
	go readLoop(conn)

	// A goroutine de escrita ROTEIA para o menu correto com base no estado.
	writeLoop(conn)
}

// --- Goroutines ---

// readLoop escuta o servidor e ATUALIZA O ESTADO DO CLIENTE.
func readLoop(conn net.Conn) {
	for {
		msg, err := network.ReadMessage(conn)
		if err != nil {
			log.Println("\nConexão com o servidor perdida. Pressione Enter para sair.")
			os.Exit(0)
			return
		}

		// *** LÓGICA CRÍTICA DE MUDANÇA DE ESTADO ***
		if msg.Type == "RESPONSE_SUCCESS" {
			var payload struct{ Message string `json:"message"` }
			json.Unmarshal(msg.Payload, &payload)
			
			// Detecta as mensagens-chave do seu servidor para mudar o estado do cliente.
			if strings.Contains(payload.Message, "matchmaking queue") {
				clientState = StateInQueue
			} else if strings.Contains(payload.Message, "Match found!") || strings.Contains(payload.Message, "The round has started!") {
				clientState = StateInMatch
			} else if strings.Contains(payload.Message, "You have returned to the lobby") {
				clientState = StateMainMenu
			}
		} else if msg.Type == "GAME_OVER" { // Um tipo de mensagem específico para fim de jogo também funcionaria.
			clientState = StateMainMenu
		}

		// Após potencialmente atualizar o estado, imprime a mensagem e o prompt correto.
		printServerMessage(msg)
	}
}

// writeLoop é um ROTEADOR que chama o handler de menu correto.
func writeLoop(conn net.Conn) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		// A cada iteração, ele verifica o estado atual e mostra o menu certo.
		switch clientState {
		case StateMainMenu:
			handleMainMenu(conn, scanner)
		case StateInQueue:
			handleInQueueMenu(conn, scanner)
		case StateInMatch:
			handleInMatchMenu(conn, scanner)
		}
	}
}

// --- Handlers de Menu ---

// handleMainMenu lida com as 8 opções do lobby.
func handleMainMenu(conn net.Conn, scanner *bufio.Scanner) {
	scanner.Scan()
	choice := scanner.Text()

	var msg network.Message
	shouldSend := true

	switch choice {
	case "1": msg.Type = "FIND_MATCH"
	case "2": msg.Type = "PURCHASE_PACKAGE"
	case "3":
		amount, err := promptForInt(scanner, "Digite a quantidade: ")
		if err != nil { fmt.Println(err); shouldSend = false; return }
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
		if err != nil { fmt.Println(err); shouldSend = false; return }
		payload, _ := json.Marshal(map[string]int{"index": index})
		msg = network.Message{Type: "REMOVE_CARD_FROM_DECK", Payload: payload}
	case "8":
		index, err := promptForInt(scanner, "Digite o índice da carta a substituir: ")
		if err != nil { fmt.Println(err); shouldSend = false; return }
		key := promptForString(scanner, "Digite a chave da nova carta: ")
		payload, _ := json.Marshal(map[string]interface{}{"index": index, "key": key})
		msg = network.Message{Type: "REPLACE_CARD_TO_DECK", Payload: payload}
	default:
		fmt.Println("Opção inválida.")
		shouldSend = false
	}

	if shouldSend {
		if err := network.WriteMessage(conn, msg); err != nil {
			log.Printf("Erro ao enviar mensagem: %v", err)
		}
	}
}

// handleInQueueMenu lida com a opção de sair da fila.
func handleInQueueMenu(conn net.Conn, scanner *bufio.Scanner) {
	scanner.Scan()
	choice := scanner.Text()
	if choice == "0" {
		msg := network.Message{Type: "LEAVE_QUEUE"}
		if err := network.WriteMessage(conn, msg); err != nil {
			log.Printf("Erro ao enviar mensagem: %v", err)
		}
	} else {
		fmt.Println("Opção inválida. Digite 0 para sair da fila.")
	}
}

// handleInMatchMenu lida com a jogada de uma carta.
func handleInMatchMenu(conn net.Conn, scanner *bufio.Scanner) {
	// Durante a partida, a mão do jogador é mostrada pelo servidor.
	// O cliente só precisa pedir o índice.
	index, err := promptForInt(scanner, "\nDigite o índice da carta para jogar: ")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Conforme o seu `match_handlers.go`, o payload deve ser `{"cardIndex": ...}`
	payload, _ := json.Marshal(map[string]int{"cardIndex": index})
	msg := network.Message{Type: "PLAY_CARD", Payload: payload}
	if err := network.WriteMessage(conn, msg); err != nil {
		log.Printf("Erro ao enviar mensagem: %v", err)
	}
}

// --- Funções de Utilidade (impressão e prompts) ---



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
		return 0, fmt.Errorf("entrada inválida. Por favor, digite um número.")
	}
	return num, nil
}

func printServerMessage(msg *network.Message) {
	var successPayload struct {
		Message string `json:"message"`
		Data    any    `json:"data"`
	}
	var errorPayload struct {
		Error string `json:"error"`
	}

	if msg.Type == "RESPONSE_SUCCESS" && json.Unmarshal(msg.Payload, &successPayload) == nil {
		fmt.Printf("%s\n", successPayload.Message)
		if successPayload.Data != nil {
			fmt.Printf("%v\n", successPayload.Data)
		}
	} else if msg.Type == "RESPONSE_ERROR" && json.Unmarshal(msg.Payload, &errorPayload) == nil {
		fmt.Printf("Status: ERRO\n")
		fmt.Printf("Mensagem: %s\n", errorPayload.Error)
	} else {
		fmt.Printf("Tipo: %s\n", msg.Type)
		fmt.Printf("Payload: %s\n", string(msg.Payload))
	}
	
	// Após imprimir, mostra um prompt para guiar o usuário, refletindo o estado atual.
	switch clientState {
	case StateMainMenu:
		fmt.Print("\n(Menu Principal) Digite uma opção: ")
	case StateInQueue:
		fmt.Print("\n(Na Fila) Digite 0 para sair: ")
	case StateInMatch:
		fmt.Print("\n(Em Jogo) Aguardando sua ação... ")
	}
}