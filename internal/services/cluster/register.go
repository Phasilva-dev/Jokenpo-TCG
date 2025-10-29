//START OF FILE jokenpo/internal/cluster/register.go
package cluster

import (
	"fmt"
	"log"

	consul "github.com/hashicorp/consul/api"
)

// ServiceRegistrar gerencia o registro contínuo de um serviço no Consul.
type ServiceRegistrar struct {
	consulManager *ConsulManager
	registration  *consul.AgentServiceRegistration
}

// NewServiceRegistrar cria um novo registrar que gerencia o ciclo de vida do registro de um serviço.
func NewServiceRegistrar(
	manager *ConsulManager,
	serviceName, advertisedHost string,
	servicePort, healthPort int,
) (*ServiceRegistrar, error) {

	if advertisedHost == "" {
		return nil, fmt.Errorf("a variável de ambiente SERVICE_ADVERTISED_HOSTNAME deve ser definida")
	}

	serviceID := fmt.Sprintf("%s-%s", serviceName, advertisedHost)

	registration := &consul.AgentServiceRegistration{
		ID:      serviceID,
		Name:    serviceName,
		Port:    servicePort,
		Address: advertisedHost,
		Check: &consul.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s:%d/health", advertisedHost, healthPort),
			Timeout:                        "5s",
			Interval:                       "10s",
			DeregisterCriticalServiceAfter: "1m",
		},
	}

	return &ServiceRegistrar{
		consulManager: manager,
		registration:  registration,
	}, nil
}

// Register tenta registrar o serviço usando o cliente Consul atual e ativo.
// Esta função é projetada para ser chamada múltiplas vezes (inclusive como callback).
func (r *ServiceRegistrar) Register() {
	client := r.consulManager.GetClient()
	if client == nil {
		log.Printf("[Registrar] Falha ao registrar '%s': cliente Consul indisponível.", r.registration.Name)
		return
	}

	err := client.Agent().ServiceRegister(r.registration)
	if err != nil {
		log.Printf("[Registrar] ERRO ao registrar serviço '%s' no Consul: %v", r.registration.Name, err)
		return
	}

	log.Printf("[Registrar] ✅ Serviço '%s' (ID: %s) registrado/atualizado no Consul.", r.registration.Name, r.registration.ID)
}

// Deregister remove o serviço do Consul. Útil para um desligamento gracioso.
func (r *ServiceRegistrar) Deregister() {
	client := r.consulManager.GetClient()
	if client == nil {
		log.Printf("[Registrar] Falha ao desregistrar '%s': cliente Consul indisponível.", r.registration.Name)
		return
	}

	log.Printf("[Registrar] Desregistrando serviço '%s' (ID: %s)...", r.registration.Name, r.registration.ID)
	err := client.Agent().ServiceDeregister(r.registration.ID)
	if err != nil {
		log.Printf("[Registrar] ERRO ao desregistrar serviço '%s': %v", r.registration.Name, err)
	}
}

//END OF FILE jokenpo/internal/cluster/register.go