package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strconv"

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

func main() {
	// conecta no servidor
	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	fmt.Println("Conectado ao servidor.")

	i := 0
	for {
		
	// envia uma mensagem de teste
	payload := map[string]string{
		"text": "Olá servidor! Req de número: " + strconv.Itoa(i),
	}
	i++
	payloadBytes, _ := json.Marshal(payload)

	msg := network.Message{
		Type:    "TEST",
		Payload: payloadBytes,
	}

	
	if err := writeMessage(conn, msg); err != nil {
		panic(err)
	}
	fmt.Println("Mensagem enviada.")
	

	// espera resposta
	resp, err := readMessage(conn)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Resposta recebida: %+v\n", resp)
	}

	/*// opcional: envia uma mensagem gigante para disparar o limite
	bigPayload := make([]byte, network.MaxMessageSize+100) // estoura os 10KB
	for i := range bigPayload {
		bigPayload[i] = 'A'
	}
	bigMsg := network.Message{
		Type:    "BIG_TEST",
		Payload: bigPayload,
	}
	fmt.Println("Enviando mensagem gigante...")
	if err := writeMessage(conn, bigMsg); err != nil {
		fmt.Println("Erro esperado (mensagem muito grande):", err)
	}*/
}
