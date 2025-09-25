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
	if err := card.InitGlobalCatalog(); err != nil {
		log.Fatalf("Falha fatal ao inicializar o catálogo de cartas: %v", err)
	}
	fmt.Println("Catálogo de cartas inicializado com sucesso.")

	// --- ETAPA 2: Configuração da Lógica do Jogo ---
	gameHandler := session.NewGameHandler()
	go gameHandler.Matchmaker().Run()
	fmt.Println("Matchmaker iniciado.")

	// --- ETAPA 3: Configuração da Camada de Rede ---
	server := network.NewServer(gameHandler)
	fmt.Println("Servidor de rede criado.")

	// --- ETAPA 4: Iniciar os Servidores ---

	// Inicia o listener TCP em sua própria goroutine para não bloquear.
	tcpAddress := "0.0.0.0:8080"
	go func() {
		if err := server.Listen(tcpAddress); err != nil {
			log.Fatalf("Falha fatal ao iniciar o servidor de rede TCP: %v", err)
		}
	}()

	/*
	// Inicia o listener UDP em sua própria goroutine.
	udpAddress := "0.0.0.0:8081"
	go func() {
		if err := server.ListenUDP(udpAddress); err != nil {
			// Usamos log.Printf aqui para não encerrar a aplicação se o UDP falhar
			log.Printf("Falha ao iniciar o servidor de rede UDP: %v", err)
		}
	}()*/

	fmt.Printf("Servidores TCP e UDP configurados para iniciar.\n")

	// Mantém a aplicação rodando indefinidamente.
	// Sem isso, a função main terminaria e o programa fecharia.
	select {}
}