package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strconv"
	"os"
	"log"
	"bufio"
	"time"

	"jokenpo/internal/network" // ajuste para o path real do seu módulo
)

// escreve uma mensagem seguindo o framing
func writeMessage(conn net.Conn, msg network.Message) error {
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	lenBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(lenBuf, uint32(len(msgBytes)))

	if _, err := conn.Write(lenBuf); err != nil {
		return err
	}
	if _, err := conn.Write(msgBytes); err != nil {
		return err
	}
	return nil
}

// lê uma mensagem seguindo o framing
func readMessage(conn net.Conn) (*network.Message, error) {
	lenBuf := make([]byte, 4)
	if _, err := io.ReadFull(conn, lenBuf); err != nil {
		return nil, err
	}
	msgLen := binary.LittleEndian.Uint32(lenBuf)

	if msgLen > network.MaxMessageSize {
		return nil, fmt.Errorf("resposta maior que o permitido (%d bytes)", msgLen)
	}

	msgBytes := make([]byte, msgLen)
	if _, err := io.ReadFull(conn, msgBytes); err != nil {
		return nil, err
	}

	var msg network.Message
	if err := json.Unmarshal(msgBytes, &msg); err != nil {
		return nil, err
	}

	return &msg, nil
}

// runInteractiveMode é para um jogador humano.
func runInteractiveMode() {
	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		log.Fatalf("Erro ao conectar: %v", err)
	}
	defer conn.Close()

	// Goroutine para ler respostas do servidor
	go func() {
		for {
			resp, err := readMessage(conn)
			if err != nil { return }
			log.Printf("\n<-- SERVIDOR: Tipo=%s, Payload=%s\n> ", resp.Type, string(resp.Payload))
		}
	}()

	log.Println("Modo Interativo. Formato: TIPO {\"json\":\"payload\"}")
	fmt.Print("> ")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		// Lógica para parsear e enviar a mensagem do usuário...
		// (pode reaproveitar a lógica do cliente de teste interativo anterior)
		fmt.Print("> ")
	}
}

// runBotMode é o nosso testador de carga automatizado.
func runBotMode() {
	log.Println("Iniciando cliente em MODO BOT.")
	serverAddr := os.Getenv("SERVER_ADDR")
	if serverAddr == "" {
		log.Fatal("SERVER_ADDR não definido para o modo bot.")
	}

	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		log.Fatalf("Erro ao conectar em %s: %v", serverAddr, err)
	}
	defer conn.Close()

	// Goroutine para logar respostas
	go func() {
		for {
			resp, err := readMessage(conn)
			if err != nil { return }
			log.Printf("Resposta recebida: Tipo=%s, Payload=%s", resp.Type, string(resp.Payload))
		}
	}()

	numRequests := 10
	for i := 0; i < numRequests; i++ {
		payload := map[string]string{"text": "Req #" + strconv.Itoa(i)}
		payloadBytes, _ := json.Marshal(payload)
		msg := network.Message{Type: "TEST", Payload: payloadBytes}

		if err := writeMessage(conn, msg); err != nil {
			log.Fatalf("Erro ao enviar mensagem: %v", err)
		}
		time.Sleep(500 * time.Millisecond)
	}

	log.Println("Enviando requisição para contagem de clientes...")
	getCountMsg := network.Message{Type: "GET_CLIENT_COUNT"}
	if err := writeMessage(conn, getCountMsg); err != nil {
		log.Fatalf("Erro ao enviar GET_CLIENT_COUNT: %v", err)
	}

	log.Println("Teste concluído. Cliente em modo de espera.")
	select {} // Bloqueia para sempre
}

func main() {
	// Pega o modo de operação da variável de ambiente
	clientMode := os.Getenv("CLIENT_MODE")
	
	if clientMode == "bot" {
		runBotMode()
	} else {
		runInteractiveMode()
	}
}
