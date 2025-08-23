package main

import (
	"encoding/json"
	"fmt"
	"net"
)

type Request struct {
	A int `json:"a"`
	B int `json:"b"`
}

type Response struct {
	Result int `json:"result"`
}

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	fmt.Println("Servidor rodando na porta 8080...")

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return
	}

	var req Request
	json.Unmarshal(buf[:n], &req)

	// Calcula soma
	result := req.A + req.B

	resp := Response{Result: result}
	data, _ := json.Marshal(resp)

	conn.Write(data)
}
