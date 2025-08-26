package network

import (
	"fmt"
	"net"
)

// Server é a estrutura principal do nosso servidor de rede.
// Agora ele gerencia um Hub.
type Server struct {
	hub *Hub
}

// NewServer agora aceita um EventHandler para passá-lo ao Hub.
// Este é o ponto de injeção da lógica do seu jogo.
func NewServer(handler EventHandler) *Server {
	return &Server{
		hub: NewHub(handler), // Cria o Hub associado a este servidor
	}
}

func (s *Server) Listen(address string) error {
	// 1. Inicia o listener na porta especificada.
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err // Não foi possível abrir a porta.
	}
	defer listener.Close() // Garante que o listener será fechado ao final.

	go s.hub.Run()

	fmt.Printf("Servidor escutando em %s\n", address)

	for {
		// listener.Accept() é uma chamada bloqueante. O código para aqui até um cliente conectar.
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Erro ao aceitar conexão: %v\n", err)
			continue // Tenta aceitar a próxima conexão.
		}

		// 4. Para cada nova conexão, criamos um novo Cliente.
		client := &Client{
			conn: conn,
			hub:  s.hub,
			send: make(chan Message, 256), // Cria um canal de envio com buffer
		}

		// 5. Registra o novo cliente no Hub.
		client.hub.register <- client

		// 6. Inicia as goroutines de leitura e escrita para este cliente.
		// Agora o cliente está totalmente ativo e independente.
		go client.writeLoop()
		go client.readLoop()
	}

}