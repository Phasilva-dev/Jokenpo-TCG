//START OF FILE jokenpo/internal/network/handler.go
package network

// EventHandler é a interface que conecta a lógica da rede com a lógica do jogo.
// O nosso código de jogo (fora deste pacote) irá implementar esta interface.
type EventHandler interface {
	// OnConnect é chamado quando um novo cliente se conecta com sucesso.
	OnConnect(c *Client)

	// OnDisconnect é chamado quando um cliente se desconecta.
	OnDisconnect(c *Client)

	// OnMessage é chamado quando uma nova mensagem é recebida de um cliente. 
	// Basicamente um handler para qualquer solicitação
	OnMessage(c *Client, msg Message)
}

//END OF FILE jokenpo/internal/network/handler.go