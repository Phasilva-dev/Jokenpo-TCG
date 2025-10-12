package cluster

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	consul "github.com/hashicorp/consul/api"
)

// Retorna um endereço "host:port" aleatório de um serviço registrado
func DiscoverService(serviceName string) string {
	config := consul.DefaultConfig()
	if config.Address == "" {
		config.Address = "consul:8500" // padrão Docker
	}

	client, err := consul.NewClient(config)
	if err != nil {
		log.Fatalf("Erro ao criar cliente Consul: %v", err)
	}

	services, _, err := client.Health().Service(serviceName, "", true, nil)
	if err != nil || len(services) == 0 {
		log.Fatalf("Nenhum serviço disponível para '%s'", serviceName)
	}

	rand.Seed(time.Now().UnixNano())
	s := services[rand.Intn(len(services))]

	return s.Service.Address + ":" + fmt.Sprint(s.Service.Port)
}