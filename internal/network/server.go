package network

import (
	"fmt"
	"io"
	"net"
)

type Server struct {

}

// NewServer cria uma nova instância do nosso servidor.
func NewServer() *Server {
	return &Server{}
}

func (s *Server) Listen(address string) error {
	// 1. Inicia o listener na porta especificada.
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err // Não foi possível abrir a porta.
	}
	defer listener.Close() // Garante que o listener será fechado ao final.

	fmt.Printf("Servidor escutando em %s\n", address)

	for {
		// listener.Accept() é uma chamada bloqueante. O código para aqui até um cliente conectar.
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Erro ao aceitar conexão: %v\n", err)
			continue // Tenta aceitar a próxima conexão.
		}

		// 3. Inicia uma NOVA GOROUTINE para cada cliente.
		// A palavra "go" é a mágica da concorrência em Go.
		// Isso permite que o loop 'for' volte imediatamente para o listener.Accept()
		// e espere por novos clientes, enquanto o cliente atual é tratado em paralelo.
		go s.handleConnection(conn)
	}

}

func (s *Server) handleConnection(conn net.Conn) {
	
	defer conn.Close()

	clientAddr := conn.RemoteAddr().String()
	fmt.Printf("Novo cliente conectado: %s\n", clientAddr)

	// Loop infinito para ler mensagens do cliente.
	for {
		// Usamos a função que criamos no protocol.go!
		msg, err := ReadMessage(conn)
		if err != nil {
			// Se o erro for io.EOF, significa que o cliente desconectou de forma limpa.
			if err == io.EOF {
				fmt.Printf("Cliente %s desconectou.\n", clientAddr)
			} else {
				fmt.Printf("Erro ao ler mensagem do cliente %s: %v\n", clientAddr, err)
			}
			return // Sai da função e a conexão é fechada pelo defer.
		}

		// Por enquanto, apenas imprimimos a mensagem recebida no console do servidor.
		fmt.Printf("Mensagem recebida de %s: Tipo=%s, Payload=%s\n", clientAddr, msg.Type, string(msg.Payload))
	}
}
