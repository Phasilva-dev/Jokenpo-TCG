package cluster

import (
	"fmt"
	"log"
	"math/rand/v2"

	consul "github.com/hashicorp/consul/api"
)

// DiscoveryMode define os diferentes tipos de descoberta de serviço.
type DiscoveryMode int

const (
	// ModeAnyHealthy retorna qualquer instância saudável do serviço (para round-robin).
	ModeAnyHealthy DiscoveryMode = iota
	// ModeLeader encontra o nó que detém o lock de liderança no Consul.
	ModeLeader
	// ModeSpecific encontra uma instância específica do serviço pelo seu Node Name (hostname).
	ModeSpecific
)

// DiscoveryOptions permite especificar como um serviço deve ser descoberto.
type DiscoveryOptions struct {
	Mode       DiscoveryMode
	SpecificID string // Usado apenas com ModeSpecific (deve corresponder a um hostname).
}

// Discover é a função unificada para encontrar serviços no Consul.
func Discover(serviceName string, consulAddr string, opts DiscoveryOptions) string {
	config := consul.DefaultConfig()
	config.Address = consulAddr

	client, err := consul.NewClient(config)
	if err != nil {
		log.Printf("ERRO: Erro ao criar cliente Consul para descoberta: %v", err)
		return ""
	}

	// Delega para a função helper apropriada com base no modo.
	switch opts.Mode {
	case ModeLeader:
		return discoverLeader(client, serviceName)
	case ModeSpecific:
		if opts.SpecificID == "" {
			log.Printf("ERRO: ModeSpecific requer um SpecificID (hostname) não vazio.")
			return ""
		}
		return discoverSpecific(client, serviceName, opts.SpecificID)
	case ModeAnyHealthy:
		fallthrough // Se o modo for AnyHealthy, executa o caso padrão.
	default:
		return discoverAnyHealthy(client, serviceName)
	}
}

// --- Funções Helper Privadas ---

// discoverLeader encontra o endereço do nó líder.
func discoverLeader(client *consul.Client, serviceName string) string {
	leaderKey := fmt.Sprintf("service/%s/leader", serviceName)
	kvPair, _, err := client.KV().Get(leaderKey, nil)
	if err != nil || kvPair == nil || len(kvPair.Value) == 0 {
		log.Printf("AVISO: Nenhum líder eleito encontrado para o serviço '%s'", serviceName)
		return ""
	}
	leaderNodeID := string(kvPair.Value)

	log.Printf("INFO: Líder para '%s' é o nó '%s'. Procurando seu endereço...", serviceName, leaderNodeID)
	// Reutiliza a lógica de busca específica para encontrar o endereço do líder.
	return discoverSpecific(client, serviceName, leaderNodeID)
}

// discoverSpecific encontra uma instância específica pelo seu ID de nó (hostname).
func discoverSpecific(client *consul.Client, serviceName string, nodeID string) string {
	services, _, err := client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		log.Printf("ERRO: Falha ao buscar instâncias do serviço '%s': %v", serviceName, err)
		return ""
	}

	for _, service := range services {
		// Compara o nodeID procurado com o nome do nó que o Consul conhece.
		if service.Service.Address == nodeID {
			addr := service.Service.Address
			port := service.Service.Port

			/*// Fallback: Se o endereço do serviço estiver vazio por algum motivo, usa o IP do nó.
			if addr == "" {
				addr = service.Node.Address
			}*/

			foundAddr := fmt.Sprintf("%s:%d", addr, port)
			log.Printf("INFO: Endereço do nó específico '%s' encontrado: %s", nodeID, foundAddr)
			return foundAddr
		}
	}

	log.Printf("AVISO: Nó específico '%s' não foi encontrado na lista de serviços saudáveis para '%s'.", nodeID, serviceName)
	return ""
}

// discoverAnyHealthy retorna um nó saudável aleatório.
func discoverAnyHealthy(client *consul.Client, serviceName string) string {
	services, _, err := client.Health().Service(serviceName, "", true, nil)
	if err != nil || len(services) == 0 {
		log.Printf("AVISO: Nenhum serviço saudável encontrado para '%s'", serviceName)
		return ""
	}

	// Escolhe um serviço aleatório da lista de saudáveis.
	s := services[rand.IntN(len(services))]

	addr := s.Service.Address
	if addr == "" {
		addr = s.Node.Address
	}

	return fmt.Sprintf("%s:%d", addr, s.Service.Port)
}