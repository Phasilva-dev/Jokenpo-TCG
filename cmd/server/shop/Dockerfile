# --- Estágio 1: Build (O Construtor) ---
# Usamos uma imagem oficial do Go (versão Alpine para ser menor) para compilar nosso código.
# NOTA: golang:1.25.0 ainda não existe, a versão mais recente é 1.22. Usei 1.22 como exemplo.
FROM golang:1.25-alpine AS builder

# Define o diretório de trabalho dentro do contêiner de build.
WORKDIR /app

# Copia os arquivos de gerenciamento de dependências primeiro.
# Isso aproveita o cache do Docker: se os módulos não mudarem, o 'go mod download' não é executado novamente.
COPY go.mod go.sum ./
RUN go mod download

# Copia todo o resto do código-fonte do projeto (o contexto é a raiz).
COPY . .

# --- MUDANÇA ---
# Compila a aplicação do Shop Service para um único binário estático.
# - CGO_ENABLED=0: Garante que o binário não tenha dependências de bibliotecas C.
# - GOOS=linux: Garante que o binário seja compilado para o sistema operacional Linux (usado nos contêineres).
# -o /shopservice: Define o nome do arquivo de saída como 'shopservice' e o coloca na raiz do sistema de arquivos.
# ./cmd/shopservice: O caminho para o pacote 'main' do nosso Shop Service.
RUN CGO_ENABLED=0 GOOS=linux go build -o /shopservice ./cmd/server/shop


# --- Estágio 2: Final (A Imagem de Produção) ---
# Usamos a imagem 'scratch', que é uma imagem completamente vazia.
# É a base mais segura e leve possível para um binário estático.
FROM scratch

# --- MUDANÇA ---
# Copia APENAS o binário compilado do estágio de build para a nossa imagem final.
# Todo o código-fonte e as ferramentas de compilação são descartados.
COPY --from=builder /shopservice /shopservice

# --- MUDANÇA ---
# Expõe a porta 8081, que é a porta que nosso Shop Service escuta.
# Isso é principalmente para documentação; a publicação da porta é feita no docker-compose.yml.
EXPOSE 8081

# --- MUDANÇA ---
# O comando que será executado quando o contêiner iniciar.
CMD ["/shopservice"]