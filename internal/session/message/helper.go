package message

import (
	"fmt"
	"jokenpo/internal/network"
)

// MessageSender define a interface para qualquer tipo que pode receber uma mensagem.
// Isso nos permite desacoplar o pacote `message` de implementações concretas como `PlayerSession` ou `network.Client`.
type MessageSender interface {
	Send() chan<- network.Message
}

// SendError envia apenas uma mensagem de erro para o cliente.
func SendError(sender MessageSender, format string, args ...interface{}) {
	errorMsg := fmt.Sprintf(format, args...)
	sender.Send() <- CreateErrorResponse(errorMsg)
}

// SendErrorAndPrompt envia uma mensagem de erro seguida por um prompt de input.
func SendErrorAndPrompt(sender MessageSender, format string, args ...interface{}) {
	errorMsg := fmt.Sprintf(format, args...)
	sender.Send() <- CreateErrorResponse(errorMsg)
	sender.Send() <- CreatePromptInputMessage()
}

// SendSuccess envia apenas uma mensagem de sucesso para o cliente.
func SendSuccess(sender MessageSender, state, message string, data any) {
	sender.Send() <- CreateSuccessResponse(state, message, data)
}

// SendSuccessAndPrompt envia uma mensagem de sucesso seguida por um prompt de input.
func SendSuccessAndPrompt(sender MessageSender, state, message string, data any) {
	sender.Send() <- CreateSuccessResponse(state, message, data)
	sender.Send() <- CreatePromptInputMessage()
}

func SendPromptInput(sender MessageSender) { //, state, message string, data any
	sender.Send() <- CreatePromptInputMessage()
}