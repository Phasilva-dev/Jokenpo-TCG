//START OF FILE jokenpo/internal/services/cluster/health.go
package cluster

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// ----------------------------------------------------------------------------
// --- Opção 1: O Handler Básico e Reutilizável ---
// ----------------------------------------------------------------------------

// NewBasicHealthHandler retorna um http.HandlerFunc genérico que qualquer serviço
// pode usar para um simples "liveness check". Ele apenas confirma que o
// processo está rodando e o servidor HTTP está respondendo.
func NewBasicHealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Service is alive.")
	}
}


// ----------------------------------------------------------------------------
// --- Opção 2: O Construtor de Handlers Avançados (Health Aggregator) ---
// ----------------------------------------------------------------------------

// CheckFunc é um tipo para uma função que realiza uma verificação de saúde.
// Retorna um erro se a verificação falhar.
type CheckFunc func() error

// HealthAggregator permite registrar múltiplas verificações de saúde e as expõe
// através de um único endpoint HTTP.
type HealthAggregator struct {
	mu     sync.RWMutex
	checks map[string]CheckFunc
}

// NewHealthAggregator cria um novo agregador de saúde.
func NewHealthAggregator() *HealthAggregator {
	return &HealthAggregator{
		checks: make(map[string]CheckFunc),
	}
}

// AddCheck registra uma nova função de verificação.
func (h *HealthAggregator) AddCheck(name string, check CheckFunc) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checks[name] = check
}

// Handler retorna um http.HandlerFunc que executa todas as verificações registradas.
// Se todas passarem, retorna 200 OK. Se alguma falhar, retorna 503 Service Unavailable.
func (h *HealthAggregator) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.mu.RLock()
		defer h.mu.RUnlock()

		errors := make(map[string]string)
		for name, check := range h.checks {
			if err := check(); err != nil {
				errors[name] = err.Error()
			}
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		if len(errors) > 0 {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(errors)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	}
}
//END OF FILE jokenpo/internal/services/cluster/health.go