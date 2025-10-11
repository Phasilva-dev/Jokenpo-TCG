# --- Estágio 1: Build (O Construtor) ---
# Usa uma imagem Go para compilar o projeto
FROM golang:1.25.0-alpine AS builder

# O diretório de trabalho dentro do contêiner de build
WORKDIR /app

# Copia os arquivos de dependência. O contexto de build (definido no docker-compose)
# será a raiz do projeto, então esses arquivos estarão disponíveis.
COPY go.mod go.sum ./
RUN go mod download

# Copia todo o código fonte da raiz do projeto para o WORKDIR
COPY . .

# Compila o binário do servidor. O caminho é relativo ao WORKDIR.
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

# --- Estágio 2: Final (A Imagem Final) ---
# Usa scratch para uma imagem mínima
FROM scratch

# Copia apenas o binário compilado do estágio de build
COPY --from=builder /server /server

# Expõe a porta do jogo e a porta do health check.
# Isso é importante para documentação.
EXPOSE 8080
EXPOSE 8000

# Comando para iniciar o servidor
CMD ["/server"]