package message
//Isso aqui são as mensagens que vão no sentido servidor -> client
import (
	"encoding/json"
	"jokenpo/internal/network"
)

// SuccessClientPayload agora carrega o estado explícito.
type SuccessClientPayload struct {
	State   string `json:"state"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// ErrorPayload define a estrutura de uma resposta de erro.
type ErrorClientPayload struct {
	Error string `json:"error"`
}

// CreateSuccessResponse agora precisa saber qual estado enviar.
func CreateSuccessResponse(state, message string, data any) network.Message {
	payload := SuccessClientPayload{
		State:   state, 
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

// CreatePromptInputMessage cria a mensagem de controle dedicada.
// Seu único trabalho é dizer ao cliente para mostrar um prompt.
func CreatePromptInputMessage() network.Message {
	return network.Message{
		Type:    "PROMPT_INPUT",
		Payload: nil, // Não precisa de payload
	}
}