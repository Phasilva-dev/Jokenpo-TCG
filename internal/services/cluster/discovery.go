package cluster

import (
	"fmt"
	"log"
	"math/rand/v2"

	consul "github.com/hashicorp/consul/api"
)

// MODIFICADO: DiscoverService agora recebe o endereço do Consul.
func DiscoverService(serviceName string, consulAddr string) string {
	config := consul.DefaultConfig()
	// Usa o endereço fornecido.
	config.Address = consulAddr

	client, err := consul.NewClient(config)
	if err != nil {
		log.Printf("ERRO: Erro ao criar cliente Consul para descoberta: %v", err)
		return "" // Retorna vazio em caso de falha, em vez de Fatal.
	}

	services, _, err := client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		log.Printf("ERRO: Falha ao buscar serviço '%s': %v", serviceName, err)
		return ""
	}
    if len(services) == 0 {
        log.Printf("AVISO: Nenhum serviço saudável encontrado para '%s'", serviceName)
		return ""
    }

	s := services[rand.IntN(len(services))]

	return s.Service.Address + ":" + fmt.Sprint(s.Service.Port)
}