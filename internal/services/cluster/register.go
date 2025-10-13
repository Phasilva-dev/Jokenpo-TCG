package cluster

import (
	"fmt"
	"log"
	"os"

	consul "github.com/hashicorp/consul/api"
)

func RegisterServiceInConsul(serviceName string, servicePort int, healthPort int, consulAddr string) error {
	// 1. Conecta-se ao Consul (sem mudanças)
	config := consul.DefaultConfig()
	config.Address = consulAddr

	consulClient, err := consul.NewClient(config)
	if err != nil {
		return fmt.Errorf("erro ao criar cliente Consul: %w", err)
	}

	// 2. Define um ID de serviço único (sem mudanças)
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("não foi possível obter o hostname: %w", err)
	}
	serviceID := fmt.Sprintf("%s-%s", serviceName, hostname)

	// 4. Monta o registro do serviço
	registration := &consul.AgentServiceRegistration{
		ID:   serviceID,
		Name: serviceName,
		Port: servicePort,
		
		// --- MUDANÇA CRÍTICA ---
		// Agora, explicitamente anunciamos o hostname do contêiner como seu endereço.
		// O DNS interno do Docker Compose garantirá que outros serviços (como o Load Balancer)
		// consigam resolver este nome para o IP correto.
		Address: hostname,

		Check: &consul.AgentServiceCheck{
			// A URL do check também usa o hostname.
			HTTP:                           fmt.Sprintf("http://%s:%d/health", hostname, healthPort),
			Timeout:                        "5s",
			Interval:                       "10s",
			DeregisterCriticalServiceAfter: "1m",
		},
	}

	// 5. Registra o serviço (sem mudanças)
	err = consulClient.Agent().ServiceRegister(registration)
	if err != nil {
		return fmt.Errorf("falha ao registrar serviço no Consul: %w", err)
	}

	log.Printf("Serviço '%s' registrado no Consul com ID: %s", serviceName, serviceID)
	return nil
}