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

func (s *Server) ListenUDP(address string) error {
	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return fmt.Errorf("falha ao resolver endereço UDP: %w", err)
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return fmt.Errorf("falha ao iniciar listener UDP: %w", err)
	}
	defer conn.Close()
	fmt.Printf("Servidor UDP de ping escutando em %s\n", address)

	// Buffer para ler os pacotes recebidos. 1500 é um tamanho seguro para pacotes de internet.
	buffer := make([]byte, 1500)

	for {
		// ReadFromUDP é bloqueante. Ele espera até um pacote chegar.
		// Ele nos dá os dados, o endereço do remetente e qualquer erro.
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Printf("Erro de leitura UDP: %v\n", err)
			continue
		}

		// Decodifica o pacote que recebemos.
		packetType, timestamp, err := DecodePingPacket(buffer[:n])
		if err != nil {
			fmt.Printf("Erro ao decodificar pacote UDP de %s: %v\n", remoteAddr, err)
			continue
		}

		// Se for um pacote de ping, respondemos com um pong.
		if packetType == PING_PACKET_TYPE {
			// Criamos um pacote de pong, mas crucialmente, usamos o timestamp ORIGINAL do ping.
			// É assim que o cliente saberá que este pong é a resposta para o seu ping.
			pongPacket := EncodePingPacket(PONG_PACKET_TYPE, timestamp)

			// Envia o pong de volta para o endereço de onde o ping veio.
			_, err := conn.WriteToUDP(pongPacket, remoteAddr)
			if err != nil {
				fmt.Printf("Erro ao enviar pong UDP para %s: %v\n", remoteAddr, err)
			}
		}
	}
}