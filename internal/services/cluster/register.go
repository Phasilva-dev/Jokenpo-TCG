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
	consulClient, err := NewConsulClient(consulAddrs)
	if err != nil {
		return fmt.Errorf("erro ao criar cliente Consul: %w", err)
	}

	// --- MUDANÇA CRUCIAL ---
	// Em vez de os.Hostname(), lemos de uma variável de ambiente.
	// Em Docker Compose, este será o nome do serviço (ex: "jokenpo-session-1").
	advertiseAddr := os.Getenv("SERVICE_ADVERTISED_HOSTNAME")
	if advertiseAddr == "" {
		// Se não estiver definido, o registro falhará, o que é o comportamento correto
		// para evitar que um serviço se anuncie com um endereço inalcançável.
		return fmt.Errorf("a variável de ambiente SERVICE_ADVERTISED_HOSTNAME deve ser definida")
	}

	// Usamos o endereço anunciado para criar um ID único e mais legível.
	serviceID := fmt.Sprintf("%s-%s", serviceName, advertiseAddr)

	registration := &consul.AgentServiceRegistration{
		ID:      serviceID,
		Name:    serviceName,
		Port:    servicePort,
		Address: advertiseAddr, // <-- USA O ENDEREÇO ANUNCIADO
		Check: &consul.AgentServiceCheck{
			// O health check também precisa usar o endereço alcançável.
			HTTP:                           fmt.Sprintf("http://%s:%d/health", advertiseAddr, healthPort), // <-- USA O ENDEREÇO ANUNCIADO
			Timeout:                        "5s",
			Interval:                       "10s",
			DeregisterCriticalServiceAfter: "1m",
		},
	}

	err = consulClient.Agent().ServiceRegister(registration)
	if err != nil {
		return fmt.Errorf("falha ao registrar serviço no Consul: %w", err)
	}

	log.Printf("Serviço '%s' registrado no Consul com ID: %s e endereço: %s", serviceName, serviceID, advertiseAddr)
	return nil
}

//END OF FILE jokenpo/internal/cluster/register.go