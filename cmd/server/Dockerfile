# --- Estágio 1: Build ---
# Usamos uma imagem oficial do Go para compilar nosso código.
FROM golang:1.24.3-alpine AS builder

# Define o diretório de trabalho.
WORKDIR /app

# Copia os arquivos de módulo primeiro para aproveitar o cache do Docker.
# O 'COPY' é relativo ao 'context' que definiremos no docker-compose.yml.
COPY go.mod go.sum ./
RUN go mod download

# Copia todo o código fonte do projeto
COPY . .

# Compila o executável do servidor de forma estática, para não depender de libs externas.
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

# --- Estágio 2: Final ---
# Usamos uma imagem 'alpine' que é minúscula.
FROM alpine:latest

# Copia APENAS o binário compilado do estágio anterior.
COPY --from=builder /server /server

# Expõe a porta que nosso servidor escuta dentro do container.
EXPOSE 8080

# Comando para executar quando o container iniciar.
ENTRYPOINT ["/server"]