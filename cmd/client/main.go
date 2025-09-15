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
	"time"
	

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

	// A goroutine de leitura agora é a única responsável por TODA a impressão.
	go readLoop(conn)

	// A goroutine principal agora SÓ lê o input e envia, nada mais.
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		userInput := scanner.Text()

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

// --- Goroutine de Leitura ---
func readLoop(conn net.Conn) {
	for {
		msg, err := network.ReadMessage(conn)
		if err != nil {
			log.Println("\nConexão com o servidor perdida.")
			os.Exit(0)
		}

		// Lógica de mudança de estado agora é LIMPA e ROBUSTA
		if msg.Type == "RESPONSE_SUCCESS" {
			var payload struct {
				State string `json:"state"` // <-- Lemos o novo campo de estado
			}
			json.Unmarshal(msg.Payload, &payload)
			
			// ATUALIZAÇÃO DIRETA! Sem 'strings.Contains'.
			// O servidor é a fonte da verdade.
			if payload.State != "" {
				switch payload.State {
				case "lobby":
					clientState = StateMainMenu
				case "in-queue":
					clientState = StateInQueue
				case "in-match":
					clientState = StateInMatch
				default:
					log.Panic("alerta: Servidor enviou um estado desconhecido")
					clientState = StateMainMenu
				}
			}
		}

		printServerMessage(msg)

		if msg.Type == "PROMPT_INPUT" {
			printPrompt()
		}
	}
}

// --- Handlers de Input (Simplificados) ---
// Eles não precisam mais chamar printPrompt() em caso de erro.
func handleMainMenuInput(conn net.Conn, scanner *bufio.Scanner, choice string) {
	var msg network.Message
	shouldSend := true

	switch choice {
	case "1": msg.Type = "FIND_MATCH"
	case "2": msg.Type = "PURCHASE_PACKAGE"
	case "3":
		amount, err := promptForInt(scanner, "Digite a quantidade: ")
		if err != nil { fmt.Println(err); shouldSend = false; return } // Apenas retorna
		payload, _ := json.Marshal(map[string]int{"amount": amount})
		msg = network.Message{Type: "PURCHASE_MULTI_PACKAGE", Payload: payload}
	// ... (outros cases permanecem os mesmos, mas sem chamar printPrompt())
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
	case "9":
		shouldSend = false // Não envie uma mensagem TCP para este comando.
		doPing("localhost:8081") // Chama nossa nova função de ping.
		printPrompt() // Mostra o prompt novamente após o ping.
	default:
		fmt.Println("Opção inválida.")
		shouldSend = false
	}

	if shouldSend {
		if err := network.WriteMessage(conn, msg); err != nil {
			log.Printf("Erro ao enviar mensagem: %v", err)
		}
	} else {
		// Se a opção for inválida, o servidor não enviará um PROMPT.
		// Então, nós mesmos imprimimos o prompt para o usuário tentar de novo.
		printPrompt()
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
		printPrompt() // Pede para o usuário tentar de novo
	}
}

func handleInMatchInput(conn net.Conn, choice string) {
	index, err := strconv.Atoi(choice)
	if err != nil {
		fmt.Println("Entrada inválida. Por favor, digite um número.")
		printPrompt() // Pede para o usuário tentar de novo
		return
	}

	payload, _ := json.Marshal(map[string]int{"cardIndex": index})
	msg := network.Message{Type: "PLAY_CARD", Payload: payload}
	if err := network.WriteMessage(conn, msg); err != nil {
		log.Printf("Erro ao enviar mensagem: %v", err)
	}
}

// --- Funções de Utilidade ---

func printServerMessage(msg *network.Message) {
    if msg.Type == "PROMPT_INPUT" {
        return // Não imprime nada para a mensagem de controle
    }
	
	var successPayload struct {
		Message string `json:"message"`
		Data    any    `json:"data"`
	}
	var errorPayload struct {
		Error string `json:"error"`
	}

	if msg.Type == "RESPONSE_SUCCESS" && json.Unmarshal(msg.Payload, &successPayload) == nil {
		fmt.Printf("\n%s\n", successPayload.Message)
		if successPayload.Data != nil {
			fmt.Printf("%v\n", successPayload.Data)
		}
	} else if msg.Type == "RESPONSE_ERROR" && json.Unmarshal(msg.Payload, &errorPayload) == nil {
		fmt.Printf("\nErro: %s\n", errorPayload.Error)
	} else {
		fmt.Printf("\nInfo (%s): %s\n", msg.Type, string(msg.Payload))
	}
}

func printPrompt() {
	switch clientState {
	case StateMainMenu:
		fmt.Print("\n--- Jokenpo Card Game (Lobby) ---\n")
		fmt.Print("1. Buscar Partida\n")
		fmt.Print("2. Comprar Pacote\n")
		fmt.Print("3. Comprar Múltiplos Pacotes\n")
		fmt.Print("4. Ver Coleção\n")
		fmt.Print("5. Ver Deck\n")
		fmt.Print("6. Adicionar Carta ao Deck\n")
		fmt.Print("7. Remover Carta do Deck\n")
		fmt.Print("8. Substituir Carta no Deck\n")
		fmt.Print("9. Medir Ping (UDP)\n")
		fmt.Print("---------------------------------\n")
		fmt.Print("\n(Lobby) Digite uma opção: ")
	case StateInQueue:
		fmt.Print("\n(Na Fila) Digite 0 para sair: ")
	case StateInMatch:
		fmt.Print("\n(Em Jogo) Digite o índice da carta para jogar: ")
	}
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

func doPing(serverAddress string) {
	// Resolve o endereço do servidor UDP.
	serverAddr, err := net.ResolveUDPAddr("udp", serverAddress)
	if err != nil {
		fmt.Printf("Erro ao resolver endereço do servidor de ping: %v\n", err)
		return
	}

	// Cria uma "conexão" UDP. Para UDP, isso não estabelece uma conexão real,
	// apenas prepara um socket para enviar e receber dados.
	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		fmt.Printf("Erro ao criar conexão UDP: %v\n", err)
		return
	}
	defer conn.Close()

	// 1. Registra o tempo de início.
	startTime := time.Now()

	// 2. Codifica e envia o pacote de ping com o timestamp atual.
	pingPacket := network.EncodePingPacket(network.PING_PACKET_TYPE, startTime.UnixNano())
	_, err = conn.Write(pingPacket)
	if err != nil {
		fmt.Printf("Erro ao enviar ping: %v\n", err)
		return
	}

	fmt.Println("Ping enviado, aguardando pong...")

	// 3. Define um timeout. Isso é CRUCIAL para UDP, pois os pacotes podem se perder!
	// Se não recebermos uma resposta em 2 segundos, desistimos.
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))

	// 4. Espera pela resposta.
	buffer := make([]byte, 9)
	n, _, err := conn.ReadFromUDP(buffer)
	if err != nil {
		// O erro mais comum aqui será um timeout.
		fmt.Printf("Erro ao receber pong: %v\n", err)
		return
	}

	// 5. Registra o tempo de chegada.
	endTime := time.Now()

	// 6. Decodifica e valida o pong.
	packetType, timestamp, err := network.DecodePingPacket(buffer[:n])
	if err != nil {
		fmt.Printf("Erro ao decodificar pong: %v\n", err)
		return
	}

	if packetType != network.PONG_PACKET_TYPE {
		fmt.Printf("Recebido pacote inesperado de tipo %x\n", packetType)
		return
	}
	if timestamp != startTime.UnixNano() {
		fmt.Println("Recebido pong de um ping antigo. Ignorando.")
		return
	}

	// 7. Calcula e exibe a latência (Round-Trip Time).
	latency := endTime.Sub(startTime)
	fmt.Printf("Pong recebido! Latência: %v\n", latency)
}