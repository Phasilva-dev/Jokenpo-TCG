//START OF FILE jokenpo/internal/session/handler_funcs.go
package session

import (
	"jokenpo/internal/network"
)

// findSessionByID é um helper para encontrar uma sessão pelo seu UUID.
func (h *GameHandler) findSessionByID(sessionID string) *PlayerSession {
	// Como o mapa de sessões usa `*network.Client` como chave, precisamos iterar.
	// Para performance em larga escala, um segundo mapa `map[string]*PlayerSession` seria melhor.
	for _, session := range h.sessionsByID {
		if session.ID == sessionID {
			return session
		}
	}
	return nil
}

func (h *GameHandler) SessionsByClient() map[*network.Client]*PlayerSession { return h.sessionsByClient }
func (h *GameHandler) SessionsByID() map[string]*PlayerSession { return h.sessionsByID }

//END OF FILE jokenpo/internal/session/handler_funcs.go