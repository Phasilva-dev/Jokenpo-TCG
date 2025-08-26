package network

import (
	"net"
	"io"
	"fmt"
)

// Client é a representação de um jogador conectado do ponto de vista do servidor.
// Ele agrupa a conexão e os canais de comunicação.
type Client struct {
	// A conexão TCP real com o jogador.
	conn net.Conn

	// Uma referência ao Hub central. O cliente usa isso para se (des)registrar.
	hub *Hub

	// Um canal bufferizado para mensagens de saída.
	// O Hub coloca as mensagens aqui, e a goroutine writeLoop do cliente as envia.
	// O buffer evita que o Hub bloqueie se o cliente estiver lento para processar.
	send chan Message
}

// Conn retorna a conexão net.Conn subjacente do cliente.
// Isso é útil para o EventHandler obter informações como o endereço IP do jogador.
func (c *Client) Conn() net.Conn {
	return c.conn
}


func (c *Client) Send() chan Message {
	return c.send
}

// readLoop bombeia mensagens da conexão TCP para o Hub para processamento central.
func (c *Client) readLoop() {
	// O 'defer' garante que a limpeza sempre acontecerá quando este loop terminar.
	defer func() {
		c.hub.unregister <- c // Envia a si mesmo para o canal de desregistro.
		c.conn.Close()        // Fecha a conexão de rede.
	}()

	for {
		// Usa nossa função de protocolo para ler uma mensagem completa.
		// Esta chamada é bloqueante; ela espera até uma mensagem chegar.
		msg, err := ReadMessage(c.conn)
		if err != nil {
			// Se o erro for io.EOF, significa que o cliente desconectou normalmente.
			// Para qualquer outro erro, a conexão está quebrada. Em ambos os casos, saímos do loop.
			if err != io.EOF {
				fmt.Printf("Erro de leitura no cliente %s: %v\n", c.conn.RemoteAddr(), err)
			}
			break // Sai do loop 'for'
		}

		// Empacota a mensagem e o cliente que a enviou.
		messageToProcess := clientMessage{
			client: c,
			msg:    *msg,
		}

		// Envia o pacote para o canal de entrada do Hub.
		c.hub.incoming <- messageToProcess
	}
}

// writeLoop bombeia mensagens do canal 'send' do cliente para a conexão TCP.
func (c *Client) writeLoop() {
	// Garante que a conexão seja fechada se este loop sair.
	defer c.conn.Close()

	// O 'for range' em um canal é uma maneira elegante de processar itens
	// até que o canal seja fechado pelo Hub (no caso de desregistro).
	for msg := range c.send {
		err := WriteMessage(c.conn, msg)
		if err != nil {
			fmt.Printf("Erro de escrita no cliente %s: %v\n", c.conn.RemoteAddr(), err)
			// Se não conseguimos escrever, a conexão está morta, então paramos.
			break
		}
	}
}