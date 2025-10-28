//START OF FILE jokenpo/internal/cluster/discovery.go
package cluster

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	consul "github.com/hashicorp/consul/api"
)

type DiscoveryMode int

const (
	ModeAnyHealthy DiscoveryMode = iota
	ModeLeader
	ModeSpecific
)

type DiscoveryOptions struct {
	Mode       DiscoveryMode
	SpecificID string
}

func Discover(serviceName string, consulAddrs string, opts DiscoveryOptions) string {
	client, err := NewConsulClient(consulAddrs)
	if err != nil {
		log.Printf("ERRO: Erro ao criar cliente Consul para descoberta: %v", err)
		return ""
	}
	return discoverWithClient(client, serviceName, opts)
}

func discoverWithClient(client *consul.Client, serviceName string, opts DiscoveryOptions) string {
	switch opts.Mode {
	case ModeLeader:
		return discoverLeader(client, serviceName)
	case ModeSpecific:
		if opts.SpecificID == "" {
			log.Printf("ERRO: ModeSpecific requer um SpecificID.")
			return ""
		}
		return discoverSpecific(client, serviceName, opts.SpecificID)
	default: // ModeAnyHealthy
		return discoverAnyHealthy(client, serviceName)
	}
}

func discoverLeader(client *consul.Client, serviceName string) string {
	leaderKey := fmt.Sprintf("service/%s/leader", serviceName)
	kvPair, _, err := client.KV().Get(leaderKey, nil)
	if err != nil || kvPair == nil || len(kvPair.Value) == 0 {
		log.Printf("AVISO: Nenhum líder eleito para '%s': %v", serviceName, err)
		return ""
	}
	leaderNodeID := string(kvPair.Value)
	return discoverSpecific(client, serviceName, leaderNodeID)
}

func discoverSpecific(client *consul.Client, serviceName string, nodeID string) string {
	services, _, err := client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		log.Printf("ERRO: Falha ao buscar serviço '%s': %v", serviceName, err)
		return ""
	}
	for _, service := range services {
		if service.Service.Address == nodeID {
			addr := service.Service.Address
			port := service.Service.Port
			return fmt.Sprintf("%s:%d", addr, port)
		}
	}
	log.Printf("AVISO: Nó '%s' não encontrado ou não está saudável para o serviço '%s'.", nodeID, serviceName)
	return ""
}

func discoverAnyHealthy(client *consul.Client, serviceName string) string {
	services, _, err := client.Health().Service(serviceName, "", true, nil)
	if err != nil || len(services) == 0 {
		log.Printf("AVISO: Nenhum serviço saudável para '%s' encontrado: %v", serviceName, err)
		return ""
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	s := services[r.Intn(len(services))]
	addr := s.Service.Address
	if addr == "" {
		addr = s.Node.Address
	}
	return fmt.Sprintf("%s:%d", addr, s.Service.Port)
}
//END OF FILE jokenpo/internal/cluster/discovery.go