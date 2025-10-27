//START OF FILE jokenpo/internal/network/server.go
package network

import (
	"fmt"
	//"net"
	"net/http" // Nova importação

	"github.com/gorilla/websocket" // Nova importação
)

// Server é a estrutura principal do nosso servidor de rede.
// Agora ele gerencia um Hub.
type Server struct {
	hub *Hub
}

// upgrader armazena as configurações para promover uma conexão HTTP para WebSocket.
var upgrader = websocket.Upgrader{
	// CheckOrigin permite controlar quais domínios podem se conectar.
	// Para desenvolvimento, retornamos 'true' para permitir qualquer origem.
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// NewServer agora aceita um EventHandler para passá-lo ao Hub.
// Este é o ponto de injeção da lógica do seu jogo.
func NewServer(handler EventHandler) *Server {
	return &Server{
		hub: NewHub(handler), // Cria o Hub associado a este servidor
	}
}

// wsHandler é o nosso novo ponto de entrada para conexões de clientes.
// Ele lida com a requisição HTTP e a promove para uma conexão WebSocket.
func (s *Server) wsHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Promove a conexão HTTP para uma conexão WebSocket persistente.
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("Erro ao fazer upgrade da conexão: %v\n", err)
		return
	}

	// 2. Cria o nosso Client, agora usando a conexão WebSocket.
	// Isso agora vai compilar, pois seu client.go já foi atualizado.
	client := &Client{
		conn: conn, // conn é do tipo *websocket.Conn
		hub:  s.hub,
		send: make(chan Message, 256),
	}

	// 3. Registra o novo cliente no Hub.
	client.hub.register <- client

	// 4. Inicia as goroutines de leitura e escrita.
	go client.writeLoop()
	go client.readLoop()
}

// Listen agora inicia um servidor HTTP e configura a rota para o WebSocket.
func (s *Server) Listen(address string) error {
	// Inicia a goroutine do Hub, exatamente como antes.
	go s.hub.Run()

	// Configura o handler para a rota "/ws". Todas as conexões WebSocket virão por aqui.
	http.HandleFunc("/ws", s.wsHandler)

	fmt.Printf("Servidor WebSocket escutando em ws://%s/ws\n", address)

	// Inicia o servidor HTTP. http.ListenAndServe é bloqueante.
	err := http.ListenAndServe(address, nil)
	if err != nil {
		return err
	}

	return nil
}

//END OF FILE jokenpo/internal/network/server.go