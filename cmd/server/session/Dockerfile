# --- Estágio 1: Build (O Construtor) ---
# Usa uma imagem Go para compilar o projeto
# NOTA: golang:1.25.0 ainda não existe, a versão mais recente é 1.22. Ajuste se necessário.
FROM golang:1.25-alpine AS builder

# O diretório de trabalho dentro do contêiner de build
WORKDIR /app

# Copia os arquivos de dependência. O contexto de build (definido no docker-compose)
# será a raiz do projeto, então esses arquivos estarão disponíveis.
COPY go.mod go.sum ./
RUN go mod download

# Copia todo o código fonte da raiz do projeto para o WORKDIR
COPY . .

# --- MUDANÇA ---
# Compila o binário do serviço de sessão, usando o caminho correto para o main.go.
# O nome do binário de saída foi atualizado para clareza.
RUN CGO_ENABLED=0 GOOS=linux go build -o /jokenpo-session ./cmd/server/session


# --- Estágio 2: Final (A Imagem Final) ---
# Usa scratch para uma imagem mínima
FROM scratch

# --- MUDANÇA ---
# Copia apenas o binário compilado do estágio de build
COPY --from=builder /jokenpo-session /jokenpo-session

# Expõe a porta do jogo e a porta do health check.
EXPOSE 8080
EXPOSE 8000

# --- MUDANÇA ---
# Comando para iniciar o servidor de sessão
CMD ["/jokenpo-session"]