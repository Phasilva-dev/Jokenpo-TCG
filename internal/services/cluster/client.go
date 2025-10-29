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
	nodes := strings.Split(addrs, ",")
	for _, node := range nodes {
		node = strings.TrimSpace(node)
		//if !strings.HasPrefix(node, "http://") && !strings.HasPrefix(node, "https://") {
		//	node = "http://" + node
		//}

		cfg := consul.DefaultConfig()
		cfg.Address = node

		client, err := consul.NewClient(cfg)
		if err != nil {
			log.Printf("[NewConsulClient] Falha ao criar cliente para %s: %v", node, err)
			continue
		}

		if _, err := client.Status().Leader(); err != nil {
			log.Printf("[NewConsulClient] %s indisponível: %v", node, err)
			continue
		}

		log.Printf("[NewConsulClient] ✅ Conectado com sucesso ao nó Consul: %s", node)
		log.Printf("[NewConsulClient] Meu nó Consul definitivo agora é %s",node)
		return client, nil
	}

	return nil, fmt.Errorf("[NewConsulClient] Nenhum nó Consul disponível em: %s", addrs)
}
//END OF FILE jokenpo/internal/cluster/client.go