package main

import (
	"log"
	"jokenpo/internal/network"
)

func main() {
	// Cria uma nova instância do nosso servidor de rede.
	server := network.NewServer()

	// Inicia o servidor na porta 8080.
	// A função Listen vai bloquear e rodar para sempre (ou até dar um erro fatal).
	err := server.Listen(":8080")
	if err != nil {
		log.Fatalf("Não foi possível iniciar o servidor: %v", err)
	}
}