//START OF FILE jokenpo/internal/cluster/register.go
package cluster

import (
	"fmt"
	"log"
	"os"

	consul "github.com/hashicorp/consul/api"
)

// RegisterServiceInConsul registra um serviço no Consul, usando um cliente resiliente.
func RegisterServiceInConsul(serviceName string, servicePort int, healthPort int, consulAddrs string) error {
	// --- MUDANÇA: Usa a nova função helper para criar um cliente resiliente ---
	// Em vez de criar um cliente simples, agora criamos um que tenta múltiplos endereços.
	consulClient, err := NewConsulClient(consulAddrs)
	if err != nil {
		// Se não conseguir se conectar a NENHUM dos agentes Consul, a inicialização falha.
		return fmt.Errorf("erro ao criar cliente Consul: %w", err)
	}

	// O resto da função permanece o mesmo.
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("não foi possível obter o hostname: %w", err)
	}
	serviceID := fmt.Sprintf("%s-%s", serviceName, hostname)

	registration := &consul.AgentServiceRegistration{
		ID:      serviceID,
		Name:    serviceName,
		Port:    servicePort,
		Address: hostname,
		Check: &consul.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s:%d/health", hostname, healthPort),
			Timeout:                        "5s",
			Interval:                       "10s",
			DeregisterCriticalServiceAfter: "1m",
		},
	}

	err = consulClient.Agent().ServiceRegister(registration)
	if err != nil {
		return fmt.Errorf("falha ao registrar serviço no Consul: %w", err)
	}

	log.Printf("Serviço '%s' registrado no Consul com ID: %s", serviceName, serviceID)
	return nil
}
//END OF FILE jokenpo/internal/cluster/register.go