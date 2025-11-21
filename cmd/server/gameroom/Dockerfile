# --- Estágio 1: Build (O Construtor) ---
# Usa uma imagem oficial do Go para compilar nosso código.
# Garanta que esta versão corresponda à do seu arquivo go.mod (ex: 1.22).
FROM golang:1.25-alpine AS builder

# Define o diretório de trabalho dentro do contêiner de build.
WORKDIR /app

# Copia os arquivos de gerenciamento de dependências primeiro para otimização de cache.
COPY go.mod go.sum ./
RUN go mod download

# Copia todo o resto do código-fonte do projeto.
COPY . .

# Compila a aplicação do GameRoom Service para um único binário estático.
# -o /gameroomservice: Define o nome do arquivo de saída.
# ./cmd/server/gameroom: O caminho para o pacote 'main' do nosso GameRoom Service.
RUN CGO_ENABLED=0 GOOS=linux go build -o /gameroomservice ./cmd/server/gameroom


# --- Estágio 2: Final (A Imagem de Produção) ---
# Usa a imagem 'scratch' para segurança e tamanho mínimo.
FROM scratch

# Copia APENAS o binário compilado do estágio de build para a nossa imagem final.
COPY --from=builder /gameroomservice /gameroomservice

# Expõe a porta 8083, que é a porta que nosso GameRoom Service escuta.
EXPOSE 8083

# O comando que será executado quando o contêiner iniciar.
CMD ["/gameroomservice"]