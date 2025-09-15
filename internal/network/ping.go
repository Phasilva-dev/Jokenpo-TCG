// internal/network/ping.go
package network

import (
	"encoding/binary"
	"fmt"
)

const (
	// Definimos "tipos" de pacotes para saber o que recebemos.
	PING_PACKET_TYPE byte = 0x01
	PONG_PACKET_TYPE byte = 0x02
)

// O nosso pacote UDP será muito simples:
// [ 1 byte para o tipo de pacote ] [ 8 bytes para o timestamp ]
// O timestamp será o tempo em nanossegundos desde a "época Unix".

// EncodePingPacket cria um pacote de 9 bytes para ser enviado.
func EncodePingPacket(packetType byte, timestamp int64) []byte {
	// Cria um buffer de 9 bytes. 1 para o tipo + 8 para o int64 do timestamp.
	buf := make([]byte, 9)
	buf[0] = packetType
	binary.LittleEndian.PutUint64(buf[1:], uint64(timestamp))
	return buf
}

// DecodePingPacket lê um pacote de 9 bytes e extrai as informações.
func DecodePingPacket(data []byte) (packetType byte, timestamp int64, err error) {
	if len(data) < 9 {
		return 0, 0, fmt.Errorf("pacote UDP muito pequeno: esperado 9 bytes, recebeu %d", len(data))
	}
	packetType = data[0]
	timestamp = int64(binary.LittleEndian.Uint64(data[1:]))
	return packetType, timestamp, nil
}