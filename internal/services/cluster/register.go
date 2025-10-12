package cluster

import (
	"log"
	"os"
	"fmt"

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

	serviceID := fmt.Sprintf("%s-%s", serviceName, os.Getenv("HOSTNAME"))

	registration := &consul.AgentServiceRegistration{
		ID:      serviceID,
		Name:    serviceName,
		Port:    servicePort,
		Address: os.Getenv("HOSTNAME"),
		Check: &consul.AgentServiceCheck{
			HTTP:     fmt.Sprintf("http://%s:%d/health", os.Getenv("HOSTNAME"), healthPort),
			Interval: "10s",
			Timeout:  "1s",
		},
	}

	err = consulClient.Agent().ServiceRegister(registration)
	if err != nil {
		log.Fatalf("Falha ao registrar serviço no Consul: %s", err)
	}

	log.Printf("Serviço '%s' registrado no Consul com ID: %s", serviceName, serviceID)
}