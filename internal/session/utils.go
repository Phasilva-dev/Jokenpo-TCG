//START OF FILE jokenpo/internal/session/utils.go
package session

import (
	"fmt"
	"os"
)

// buildCallbackURL é um pequeno helper para construir URLs de callback.
func (h *GameHandler) buildCallbackURL(session *PlayerSession, path string) string {
	// O hostname do contêiner onde esta sessão está rodando.
	hostname, _ := os.Hostname()
	// A porta deste serviço jokenpo-session.
	port := 8080 // Ou obtido da configuração
	return fmt.Sprintf("http://%s:%d%s", hostname, port, path)
}

//END OF FILE jokenpo/internal/session/utils.go