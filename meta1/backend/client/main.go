package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"time"
)

// Estruturas de request/response
type Request struct {
	A int `json:"a"`
	B int `json:"b"`
}

type Response struct {
	Result int `json:"result"`
}

func main() {
	// Flags
	a := flag.Int("a", 0, "primeiro número")
	b := flag.Int("b", 0, "segundo número")
	flag.Parse()

	// Conectar ao servidor
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Monta request
	req := Request{A: *a, B: *b}
	data, _ := json.Marshal(req)

	// Marca tempo antes de enviar
	start := time.Now()

	// Envia
	_, err = conn.Write(data)
	if err != nil {
		panic(err)
	}

	// Lê resposta
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		panic(err)
	}

	// Marca tempo após receber
	elapsed := time.Since(start)

	// Decodifica resposta
	var resp Response
	json.Unmarshal(buf[:n], &resp)

	fmt.Printf("Resultado = %d | Ping = %v\n", resp.Result, elapsed)
}
