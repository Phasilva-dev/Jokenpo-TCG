//START OF FILE jokenpo/internal/network/protocol.go
package network

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"net"
	"fmt"
)

// Message é o envelope padrão para toda a comunicação.
// Ele contém um tipo para roteamento e um payload com os dados.
// As structs tag como json:"type" serve para manter a convenção de cada linguagem
type Message struct {
	Type    string          `json:"type"`    // Ex: "PLAY_CARD", "GAME_STATE_UPDATE", 
	Payload json.RawMessage `json:"payload"` // Dados específicos, mantidos em formato JSON bruto para decodificação posterior.
}

const MaxMessageSize = 1024 * 1024 // 1000 Kilobytes

func WriteMessage(conn net.Conn, msg Message) error {
	
	// Transforma nossa struct em Json, se ligue nas tags que você mesmo definiu
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// O "aperto de mão secreto" começo da definição do Framing, algo como dizer qual será o tamanho de uma mensagem pra outra
	lenBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(lenBuf, uint32(len(msgBytes)))

	// 3. Escrever o tamanho da mensagem (4 bytes) na conexão.
	if _, err := conn.Write(lenBuf); err != nil {
		return err // Erro de rede ao tentar escrever o tamanho.
	}

	// 4. Escrever a mensagem real (os bytes do JSON) na conexão.
	if _, err := conn.Write(msgBytes); err != nil {
		return err // Erro de rede ao tentar escrever o corpo da mensagem.
	}

	return nil
	

}

// ReadMessage lê uma mensagem da conexão, tratando o prefixo de tamanho.
// Esta função implementa o nosso protocolo de "framing".
func ReadMessage(conn net.Conn) (*Message, error) {
	// 1. Ler os primeiros 4 bytes para obter o tamanho da mensagem que está por vir.
	lenBuf := make([]byte, 4)
	// io.ReadFull é crucial aqui. Ele garante que ou lemos 4 bytes ou recebemos um erro.
	// Um erro comum aqui é io.EOF, que significa que o cliente desconectou.
	if _, err := io.ReadFull(conn, lenBuf); err != nil {
		return nil, err
	}

	// 2. Converter os 4 bytes lidos de volta para um inteiro (uint32).
	msgLen := binary.LittleEndian.Uint32(lenBuf)

	// ---> VERIFICAÇÃO DE SEGURANÇA <---
	// Se o cliente diz que vai enviar uma mensagem maior que nosso limite,
	// é um comportamento suspeito. Rejeitamos e fechamos a conexão.
	if msgLen > MaxMessageSize {
		return nil, fmt.Errorf("message too large: size %d exceeds max size %d", msgLen, MaxMessageSize)
	}

	// 3. Criar um buffer com o tamanho exato da mensagem e ler os dados da conexão.
	msgBytes := make([]byte, msgLen)
	if _, err := io.ReadFull(conn, msgBytes); err != nil {
		return nil, err
	}

	// 4. Deserializar os bytes do JSON de volta para a nossa struct Message.
	var msg Message
	if err := json.Unmarshal(msgBytes, &msg); err != nil {
		return nil, err // Os dados recebidos não são um JSON válido.
	}

	return &msg, nil
}

//END OF FILE jokenpo/internal/network/protocol.go