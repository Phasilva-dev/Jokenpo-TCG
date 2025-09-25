package network

import (
	"net"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Tempo para aguardar por uma escrita na conexão.
	writeWait = 10 * time.Second

	// Tempo máximo para aguardar por uma resposta de pong do cliente.
	pongWait = 60 * time.Second

	// Frequência com que enviamos pings para o cliente. Deve ser menor que pongWait.
	pingPeriod = (pongWait * 9) / 10
)

// Client é a representação de um jogador conectado do ponto de vista do servidor.
// Ele agrupa a conexão e os canais de comunicação.
type Client struct {
	// A conexão TCP real com o jogador.
	conn *websocket.Conn

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
	return c.conn.UnderlyingConn()
}


func (c *Client) Send() chan <- Message {
	return c.send
}

func (c *Client) readLoop() {
	// Garante que a limpeza ocorrerá quando o loop terminar.
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	// Configura um deadline para a próxima mensagem de pong.
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	// Configura um handler para as mensagens de pong recebidas.
	// O handler atualiza o read deadline, mantendo a conexão viva.
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// Loop infinito para ler mensagens do cliente.
	for {
		// gorilla/websocket oferece métodos para ler tipos de mensagem específicos.
		// Para nossa arquitetura, ReadJSON é o mais conveniente.
		var msg Message
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			// websocket.IsUnexpectedCloseError é útil para logar erros de desconexão inesperados.
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("Erro inesperado no cliente %s: %v\n", c.conn.RemoteAddr(), err)
			}
			// Para qualquer erro (desconexão normal ou anormal), saímos do loop.
			break
		}

		// Empacota a mensagem e o cliente que a enviou.
		messageToProcess := clientMessage{
			client: c,
			msg:    msg,
		}

		// Envia o pacote para o canal de entrada do Hub.
		c.hub.incoming <- messageToProcess
	}
}

// writeLoop bombeia mensagens do canal 'send' do cliente para a conexão WebSocket.
func (c *Client) writeLoop() {
	// Ticker para enviar pings periódicos para o cliente.
	ticker := time.NewTicker(pingPeriod)

	// Garante que a limpeza ocorrerá quando o loop terminar.
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			// Configura um deadline para a escrita para evitar bloqueios indefinidos.
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			// O canal 'send' foi fechado pelo Hub, o que significa que o cliente foi desregistrado.
			if !ok {
				// Envia uma mensagem de fechamento para o cliente.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			
			// Usa WriteJSON para enviar a struct Message como JSON.
			// A biblioteca cuida de toda a serialização e framing.
			err := c.conn.WriteJSON(msg)
			if err != nil {
				fmt.Printf("Erro de escrita no cliente %s: %v\n", c.conn.RemoteAddr(), err)
				return // Se a escrita falhar, encerramos a goroutine.
			}

		case <-ticker.C:
			// Envia uma mensagem de ping para o cliente.
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return // Se o ping falhar, a conexão está morta.
			}
		}
	}
}