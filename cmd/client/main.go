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
	"strconv" // Usaremos para converter string para int
	"strings"
)

// Definimos os estados do cliente para sabermos qual menu mostrar.
const (
	StateMainMenu = "MainMenu"
	StateInQueue  = "InQueue"
	StateInMatch  = "InMatch"
)

var clientState = StateMainMenu // O cliente começa no menu principal.

func main() {
	// ... (código de conexão igual ao anterior)
	address := "localhost:8080"
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatalf("Não foi possível conectar ao servidor em %s: %v", address, err)
	}
	defer conn.Close()
	fmt.Printf("Conectado a %s!\n", address)

	go readLoop(conn)
	writeLoop(conn)
}

// readLoop agora tem uma responsabilidade a mais: mudar o estado do cliente.
func readLoop(conn net.Conn) {
	for {
		msg, err := network.ReadMessage(conn)
		if err != nil {
			log.Println("Conexão com o servidor perdida. Pressione Enter para sair.")
			// Encerra a aplicação se a conexão cair
			os.Exit(0)
			return
		}

		// --- MUDANÇA DE ESTADO DO CLIENTE ---
		// O cliente "ouve" as mensagens do servidor para mudar seu próprio estado.
		if msg.Type == "RESPONSE_SUCCESS" {
			var payload struct{ Message string `json:"message"` }
			json.Unmarshal(msg.Payload, &payload)
			
			// Detectamos palavras-chave na mensagem do servidor.
			if strings.Contains(payload.Message, "matchmaking queue") {
				clientState = StateInQueue
			} else if strings.Contains(payload.Message, "Match found!") {
				clientState = StateInMatch
			} else if strings.Contains(payload.Message, "You have returned to the lobby") {
				clientState = StateMainMenu
			}
		}

		// Imprime a mensagem formatada (código anterior)
		printServerMessage(msg)
	}
}

// Extraí a lógica de impressão para uma função separada para manter o readLoop limpo.
func printServerMessage(msg *network.Message) {
	fmt.Println("\n<--- MENSAGEM DO SERVIDOR ---")
	var successPayload struct {
		Message string `json:"message"`
		Data    any    `json:"data"`
	}
	var errorPayload struct {
		Error string `json:"error"`
	}

	if msg.Type == "RESPONSE_SUCCESS" && json.Unmarshal(msg.Payload, &successPayload) == nil {
		fmt.Printf("Status: SUCESSO\n")
		fmt.Printf("Mensagem: %s\n", successPayload.Message)
		if successPayload.Data != nil {
			dataStr, _ := json.MarshalIndent(successPayload.Data, "", "  ")
			fmt.Printf("Dados:\n%s\n", string(dataStr))
		}
	} else if msg.Type == "RESPONSE_ERROR" && json.Unmarshal(msg.Payload, &errorPayload) == nil {
		fmt.Printf("Status: ERRO\n")
		fmt.Printf("Mensagem: %s\n", errorPayload.Error)
	} else {
		fmt.Printf("Tipo: %s\n", msg.Type)
		fmt.Printf("Payload: %s\n", string(msg.Payload))
	}
	fmt.Println("<--------------------------->")
}

// writeLoop agora é um roteador de estado.
func writeLoop(conn net.Conn) {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		// O loop principal agora decide qual menu mostrar com base no estado.
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

// handleMainMenu mostra as opções do lobby.
func handleMainMenu(conn net.Conn, scanner *bufio.Scanner) {
	fmt.Println("\n--- Menu Principal ---")
	fmt.Println("1. Procurar Partida")
	fmt.Println("2. Ver Coleção de Cartas")
	fmt.Println("3. Ver Baralho de Jogo")
	fmt.Println("4. Comprar Pacote de Cartas")
	fmt.Println("5. Adicionar Carta ao Baralho")
	fmt.Println("6. Remover Carta do Baralho")
	fmt.Println("7. Substituir Carta no Baralho")
	fmt.Println("--------------------")
	fmt.Print("Escolha uma opção: ")

	scanner.Scan()
	choice := scanner.Text()

	var msg network.Message
	switch choice {
	case "1":
		msg.Type = "FIND_MATCH"
	case "2":
		msg.Type = "VIEW_COLLECTION"
	case "3":
		msg.Type = "VIEW_DECK"
	case "4":
		msg.Type = "PURCHASE_PACKAGE"
	case "5":
		fmt.Print("Digite a chave da carta (ex: rock:5:red): ")
		scanner.Scan()
		key := scanner.Text()
		payload, _ := json.Marshal(map[string]string{"key": key})
		msg = network.Message{Type: "ADD_CARD_TO_DECK", Payload: payload}
	// Adicione os casos 6 e 7 de forma similar, pedindo o input necessário.
	case "6":
		
	default:
		fmt.Println("Opção inválida.")
		return // Volta ao início do loop para mostrar o menu novamente.
	}
	
	if err := network.WriteMessage(conn, msg); err != nil {
		log.Printf("Erro ao enviar mensagem: %v", err)
	}
}

// handleInQueueMenu mostra as opções enquanto o jogador está na fila.
func handleInQueueMenu(conn net.Conn, scanner *bufio.Scanner) {
	fmt.Println("\n--- Fila de Espera ---")
	fmt.Println("Aguardando o servidor encontrar um oponente...")
	fmt.Println("0. Sair da Fila")
	fmt.Println("--------------------")
	fmt.Print("Escolha uma opção: ")

	scanner.Scan()
	choice := scanner.Text()

	if choice == "0" {
		msg := network.Message{Type: "LEAVE_QUEUE"}
		if err := network.WriteMessage(conn, msg); err != nil {
			log.Printf("Erro ao enviar mensagem: %v", err)
		}
	} else {
		fmt.Println("Opção inválida.")
	}
}

// handleInMatchMenu mostra as opções durante a partida.
func handleInMatchMenu(conn net.Conn, scanner *bufio.Scanner) {
	fmt.Println("\n--- Em Partida ---")
	fmt.Println("Sua vez de jogar!")
	fmt.Print("Digite o número da carta na sua mão para jogar: ")

	scanner.Scan()
	choiceStr := scanner.Text()
	
	// Converte a escolha (string) para um número (int).
	cardIndex, err := strconv.Atoi(choiceStr)
	if err != nil {
		fmt.Println("Entrada inválida. Por favor, digite um número.")
		return
	}

	// Monta o payload JSON que o servidor espera.
	payload, _ := json.Marshal(map[string]int{"cardIndex": cardIndex})
	msg := network.Message{Type: "PLAY_CARD", Payload: payload}

	if err := network.WriteMessage(conn, msg); err != nil {
		log.Printf("Erro ao enviar mensagem: %v", err)
	}
}