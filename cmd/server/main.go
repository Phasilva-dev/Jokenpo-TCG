package main

import (
	"fmt"
	"log"

	"jokenpo/internal/game/card"
	"jokenpo/internal/network"
	"jokenpo/internal/session"
)

func main() {
	// --- ETAPA 1: Inicialização Global ---
	// Esta é a primeira coisa que fazemos. Se o catálogo de cartas não puder
	// ser carregado, a aplicação não pode continuar.
	if err := card.InitGlobalCatalog(); err != nil {
		log.Fatalf("Falha fatal ao inicializar o catálogo de cartas: %v", err)
	}
	fmt.Println("Catálogo de cartas inicializado com sucesso.")

	// --- ETAPA 2: Configuração da Lógica do Jogo ---
	// Criamos a instância principal que gerenciará toda a lógica do jogo.
	gameHandler := session.NewGameHandler()

	// Inicia o matchmaker em sua própria goroutine.
	go gameHandler.Matchmaker().Run()
	fmt.Println("Matchmaker iniciado.")

	// --- ETAPA 3: Configuração da Camada de Rede ---
	// Criamos o servidor de rede e injetamos nosso gameHandler.
	// A partir deste ponto, o servidor de rede notificará o gameHandler sobre
	// conexões, desconexões e mensagens.
	server := network.NewServer(gameHandler)
	fmt.Println("Servidor de rede criado.")

	// --- ETAPA 4: Iniciar o Servidor ---
	// Define o endereço e a porta em que o servidor irá escutar.
	address := "0.0.0.0:8080"
	
	// server.Listen() é uma chamada bloqueante. O programa ficará "preso" aqui,
	// aceitando novas conexões indefinidamente, até que o processo seja encerrado.
	if err := server.Listen(address); err != nil {
		log.Fatalf("Falha fatal ao iniciar o servidor de rede: %v", err)
	}
}