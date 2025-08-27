package network

import (

)

// clientMessage é uma estrutura para empacotar uma mensagem com o cliente que a enviou.
// O Hub precisa de ambos para passar para o EventHandler.
type clientMessage struct {
	client *Client
	msg    Message
}

// Hub mantém o conjunto de clientes ativos e roteia eventos para o handler.
type Hub struct {
	// Clientes registrados. O mapa de *Client para bool é um "set" em Go.
	// Acessado SOMENTE pela goroutine do Hub.
	clients map[*Client]bool

	// Canal para registrar novos clientes.
	register chan *Client

	// Canal para desregistrar clientes.
	unregister chan *Client

	// Canal para mensagens de entrada dos clientes.
	// As goroutines readLoop dos clientes enviam mensagens para este canal.
	incoming chan clientMessage

	// O handler da lógica do jogo que processará os eventos.
	handler EventHandler
}

// NewHub cria, inicializa e retorna um novo Hub.
func NewHub(handler EventHandler) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		incoming:   make(chan clientMessage),
		handler:    handler,
	}
}

func (h *Hub) Run() {

	for {
		select{
		case client := <- h.register:

			// Adiciona o cliente ao nosso mapa de registros.
			h.clients[client] = true
			// Notifica o handler da lógica do jogo que um novo cliente chegou.
			h.handler.OnConnect(client)

		case client := <- h.unregister:

			// Verifica se o cliente realmente está no nosso registro.
			if _, ok := h.clients[client]; ok {
				// Remove o cliente do mapa.
				delete(h.clients, client)
				// Fecha o canal 'send' do cliente. Isso é MUITO IMPORTANTE.
				// É o sinal para a goroutine writeLoop daquele cliente parar.
				close(client.send)
				// Notifica o handler da lógica do jogo que o cliente saiu.
				h.handler.OnDisconnect(client)
			}
			// --- Caso 3: Uma nova mensagem chegou de um cliente ---
		case clientMsg := <- h.incoming:
			// O Hub não se importa com o conteúdo da mensagem.
			// Ele simplesmente a delega para o handler da lógica do jogo processar.
			h.handler.OnMessage(clientMsg.client, clientMsg.msg)

		}
	}

}