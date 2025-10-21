# --- Estágio 1: Build (O Construtor) ---
# Usa uma imagem oficial do Go (versão Alpine para ser menor) para compilar nosso código.
# Garanta que esta versão corresponda à do seu arquivo go.mod.
FROM golang:1.25.0-alpine AS builder

# Define o diretório de trabalho dentro do contêiner de build.
WORKDIR /app

# Copia os arquivos de gerenciamento de dependências primeiro para otimização de cache.
COPY go.mod go.sum ./
RUN go mod download

# Copia todo o resto do código-fonte do projeto.
COPY . .

# Compila a aplicação do Queue Service para um único binário estático.
# -o /queueservice: Define o nome do arquivo de saída.
# ./cmd/server/queue: O caminho para o pacote 'main' do nosso Queue Service.
RUN CGO_ENABLED=0 GOOS=linux go build -o /queueservice ./cmd/server/queue


# --- Estágio 2: Final (A Imagem de Produção) ---
# Usa a imagem 'scratch', que é uma imagem completamente vazia, para segurança e tamanho mínimo.
FROM scratch

# Copia APENAS o binário compilado do estágio de build para a nossa imagem final.
COPY --from=builder /queueservice /queueservice

# Expõe a porta 8082, que é a porta que nosso Queue Service escuta (conforme definido no main.go).
# Isso serve como documentação para quem for usar a imagem.
EXPOSE 8082

# O comando que será executado quando o contêiner iniciar.
CMD ["/queueservice"]