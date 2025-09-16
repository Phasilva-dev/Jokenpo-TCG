# --- Estágio de Build ---
# Usa uma imagem Go completa para compilar o projeto
FROM golang:1.24.3-alpine AS builder

WORKDIR /app

# Copia e baixa as dependências primeiro para aproveitar o cache
COPY ../../go.mod ../../go.sum ./
RUN go mod download

# Copia todo o código fonte
COPY ../../ .

# Compila o servidor como um binário estático para Linux
# Esta é a sua excelente sugestão em ação.
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server/main.go

# --- Estágio Final ---
# Usa a imagem 'scratch', que é completamente vazia
FROM scratch

# Copia apenas o binário estático do estágio de build
COPY --from=builder /server /server

# Expõe as portas
EXPOSE 8080/tcp
EXPOSE 8081/udp

# Comando para iniciar o servidor
CMD ["/server"]