//START OF FILE jokenpo/internal/session/utils.go
package session

import (
	"fmt"
)

// buildCallbackURL é um pequeno helper para construir URLs de callback.
func (h *GameHandler) buildCallbackURL(session *PlayerSession, path string) string {
	// A porta deste serviço jokenpo-session. Assumindo 8080.
	port := 8080 
	
	// --- MUDANÇA CRUCIAL AQUI ---
	// Usa o hostname que foi passado na configuração, não o do sistema operacional.
	return fmt.Sprintf("http://%s:%d%s", h.advertisedHostname, port, path)
}

//END OF FILE jokenpo/internal/session/utils.go