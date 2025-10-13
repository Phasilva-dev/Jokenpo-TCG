package cluster

import (
	"fmt"
	"log"
	"os"

	consul "github.com/hashicorp/consul/api"
)

func RegisterServiceInConsul(serviceName string, servicePort int, healthPort int) {
	config := consul.DefaultConfig()
	config.Address = os.Getenv("CONSUL_HTTP_ADDR")
	if config.Address == "" {
		config.Address = "consul:8500"
	}

	consulClient, err := consul.NewClient(config)
	if err != nil {
		log.Fatalf("Erro ao criar cliente Consul: %s", err)
	}

	// O hostname ainda é perfeito para criar um ID de serviço único.
	hostname := os.Getenv("HOSTNAME")
	if hostname == "" {
		// Fallback caso a variável de ambiente não esteja setada
		hostname, _ = os.Hostname()
	}
	serviceID := fmt.Sprintf("%s-%s", serviceName, hostname)

	registration := &consul.AgentServiceRegistration{
		ID:   serviceID,
		Name: serviceName,
		Port: servicePort,

		// --- MUDANÇA PRINCIPAL ---
		// Comente ou remova a linha 'Address'. O agente do Consul irá
		// usar automaticamente o endereço IP do contêiner que está fazendo o registro.
		// Address: os.Getenv("HOSTNAME"),

		Check: &consul.AgentServiceCheck{
			// --- MUDANÇA SECUNDÁRIA ---
			// A URL do check ainda precisa de um host. Como o Docker Compose garante que
			// o hostname do contêiner é resolvível por DNS dentro da rede, usar o
			// hostname aqui é a abordagem correta e mais legível.
			HTTP: fmt.Sprintf("http://%s:%d/health", hostname, healthPort),

			// Aumenta o timeout para dar mais margem em ambientes de dev
			Timeout: "5s",
			Interval: "10s",
			// Boa prática: desregistra automaticamente o serviço se ele ficar
			// em estado crítico por mais de 1 minuto.
			DeregisterCriticalServiceAfter: "1m",
		},
	}

	err = consulClient.Agent().ServiceRegister(registration)
	if err != nil {
		log.Fatalf("Falha ao registrar serviço no Consul: %s", err)
	}

	log.Printf("Serviço '%s' registrado no Consul com ID: %s", serviceName, serviceID)
}