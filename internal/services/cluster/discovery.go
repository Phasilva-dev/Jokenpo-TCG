//START OF FILE jokenpo/internal/cluster/discovery.go
package cluster

import (
	"fmt"
	"log"
	"math/rand/v2"

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

// Discover agora usa o cliente resiliente para falar com o Consul.
func Discover(serviceName string, consulAddrs string, opts DiscoveryOptions) string {
	// --- MUDANÇA: Usa a nova função helper para criar um cliente resiliente ---
	client, err := NewConsulClient(consulAddrs)
	if err != nil {
		log.Printf("ERRO: Erro ao criar cliente Consul para descoberta: %v", err)
		return ""
	}

	// O resto da função permanece o mesmo.
	switch opts.Mode {
	case ModeLeader:
		log.Printf("AVISO: Nó Lider '%s' foi encontrado...", serviceName)
		return discoverLeader(client, serviceName)
	case ModeSpecific:
		if opts.SpecificID == "" {
			log.Printf("ERRO: ModeSpecific requer um SpecificID.")
			return ""
		}
		log.Printf("AVISO: Nó especifico '%s' com ID %s foi encontrado...", serviceName, opts.SpecificID)
		return discoverSpecific(client, serviceName, opts.SpecificID)
	default:
		log.Printf("AVISO: Nó qualquer saudavel '%s' foi encontrado...", serviceName)
		return discoverAnyHealthy(client, serviceName)
	}
}
// (O resto do arquivo discovery.go - discoverLeader, etc. - não precisa de mudanças)
func discoverLeader(client *consul.Client, serviceName string) string {
	leaderKey := fmt.Sprintf("service/%s/leader", serviceName)
	kvPair, _, err := client.KV().Get(leaderKey, nil)
	if err != nil || kvPair == nil || len(kvPair.Value) == 0 {
		log.Printf("AVISO: Nenhum líder eleito para '%s'", serviceName)
		return ""
	}
	leaderNodeID := string(kvPair.Value)
	log.Printf("INFO: Líder para '%s' é '%s'. Buscando endereço...", serviceName, leaderNodeID)
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
			foundAddr := fmt.Sprintf("%s:%d", addr, port)
			log.Printf("INFO: Endereço para nó '%s' encontrado: %s", nodeID, foundAddr)
			return foundAddr
		}
	}
	log.Printf("AVISO: Nó '%s' não encontrado para serviço '%s'.", nodeID, serviceName)
	return ""
}

func discoverAnyHealthy(client *consul.Client, serviceName string) string {
	services, _, err := client.Health().Service(serviceName, "", true, nil)
	if err != nil || len(services) == 0 {
		log.Printf("AVISO: Nenhum serviço saudável para '%s'", serviceName)
		return ""
	}
	s := services[rand.IntN(len(services))]
	addr := s.Service.Address
	if addr == "" {
		addr = s.Node.Address
	}
	return fmt.Sprintf("%s:%d", addr, s.Service.Port)
}
//END OF FILE jokenpo/internal/cluster/discovery.go