package game
/*
import (
    "fmt"
    "jokenpo/internal/network"
)

type GameHandler struct {
    rooms map[string]*Room
}

func NewGameHandler() *GameHandler {
    return &GameHandler{
        rooms: make(map[string]*Room),
    }
}

// Chamado quando um jogador conecta
func (g *GameHandler) OnConnect(c *network.Client) {
    fmt.Println("Novo jogador conectado:", c.Conn().RemoteAddr())
}

// Chamado quando desconecta
func (g *GameHandler) OnDisconnect(c *network.Client) {
    fmt.Println("Jogador saiu:", c.Conn().RemoteAddr())
}

// Chamado quando chega mensagem
func (g *GameHandler) OnMessage(c *network.Client, msg network.Message) {
    fmt.Println("Mensagem recebida:", msg.Type, string(msg.Payload))

    switch msg.Type {
    case "CREATE_ROOM":
        g.createRoom(c, msg.Payload)
    case "JOIN_ROOM":
        g.joinRoom(c, msg.Payload)
    case "PLAY_CARD":
        g.playCard(c, msg.Payload)
    }
}*/