//START OF FILE jokenpo/internal/cluster/client.go
package cluster

import (
	"fmt"
	"log"
	"strings"

	consul "github.com/hashicorp/consul/api"
)

// NewConsulClient cria um novo cliente Consul, tentando se conectar a uma lista
// de endereços fornecidos até encontrar um agente saudável com um líder.
// Isso torna a conexão inicial com o cluster Consul de alta disponibilidade.
func NewConsulClient(addrs string) (*consul.Client, error) {
	if addrs == "" {
		return nil, fmt.Errorf("endereço do Consul não pode ser vazio")
	}
	
	// Itera sobre a lista de endereços separados por vírgula.
	for _, addr := range strings.Split(addrs, ",") {
		config := consul.DefaultConfig()
		config.Address = strings.TrimSpace(addr)
		
		client, err := consul.NewClient(config)
		if err == nil {
			// Tenta fazer uma chamada simples para verificar se o agente está vivo e o cluster tem um líder.
			if _, err := client.Status().Leader(); err == nil {
				log.Printf("Conectado com sucesso ao agente Consul em %s", config.Address)
				return client, nil
			}
		}
		log.Printf("AVISO: Falha ao conectar ou verificar o agente Consul em %s: %v", config.Address, err)
	}
	
	// Se o loop terminar, não conseguimos nos conectar a nenhum agente.
	return nil, fmt.Errorf("não foi possível conectar a nenhum dos agentes Consul fornecidos: %s", addrs)
}
//END OF FILE jokenpo/internal/cluster/client.go