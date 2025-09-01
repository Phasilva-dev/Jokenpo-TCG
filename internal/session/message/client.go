package message
//Isso aqui são as mensagens que vão no sentido servidor -> client
import (
	"encoding/json"
	"jokenpo/internal/network"
)

// SuccessPayload define a estrutura de uma resposta de sucesso.
// Usar uma struct evita erros de digitação em chaves de mapa ("mesage" vs "message").
type SuccessClientPayload struct {
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"` // omitempty remove o campo se Data for nil
}

// ErrorPayload define a estrutura de uma resposta de erro.
type ErrorClientPayload struct {
	Error string `json:"error"`
}

// CreateSuccessResponse usando a struct
func CreateSuccessResponse(message string, data any) network.Message {
	payload := SuccessClientPayload{
		Message: message,
		Data:    data,
	}
	payloadBytes, _ := json.Marshal(payload)
	return network.Message{
		Type:    "RESPONSE_SUCCESS",
		Payload: payloadBytes,
	}
}

// CreateErrorResponse usando a struct
func CreateErrorResponse(errorMsg string) network.Message {
	payload := ErrorClientPayload{
		Error: errorMsg,
	}
	payloadBytes, _ := json.Marshal(payload)
	return network.Message{
		Type:    "RESPONSE_ERROR",
		Payload: payloadBytes,
	}
}