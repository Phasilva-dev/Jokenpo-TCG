# --- Estágio 1: Build (O Construtor) ---
# Usamos uma imagem oficial do Go (versão Alpine para ser menor) para compilar nosso código.
FROM golang:1.25.0-alpine AS builder

# Define o diretório de trabalho dentro do contêiner de build.
WORKDIR /app

# Copia os arquivos de gerenciamento de dependências primeiro.
# Isso aproveita o cache do Docker: se os módulos não mudarem, o 'go mod download' não é executado novamente.
COPY go.mod go.sum ./
RUN go mod download

# Copia todo o resto do código-fonte do projeto (o contexto é a raiz).
COPY . .

# Compila a aplicação do Load Balancer para um único binário estático.
# - CGO_ENABLED=0: Garante que o binário não tenha dependências de bibliotecas C.
# - GOOS=linux: Garante que o binário seja compilado para o sistema operacional Linux (usado nos contêineres).
# -o /lb: Define o nome do arquivo de saída como 'lb' e o coloca na raiz do sistema de arquivos.
# ./cmd/loadbalancer: O caminho para o pacote 'main' do nosso LB.
RUN CGO_ENABLED=0 GOOS=linux go build -o /lb ./cmd/loadbalancer

# --- Estágio 2: Final (A Imagem de Produção) ---
# Usamos a imagem 'scratch', que é uma imagem completamente vazia.
# É a base mais segura e leve possível para um binário estático.
FROM scratch

# Copia APENAS o binário compilado do estágio de build para a nossa imagem final.
# Todo o código-fonte e as ferramentas de compilação são descartados.
COPY --from=builder /lb /lb

# Expõe a porta 80, que é a porta que nosso LB escuta.
# Isso é principalmente para documentação; a publicação da porta é feita no docker-compose.yml.
EXPOSE 80

# O comando que será executado quando o contêiner iniciar.
CMD ["/lb"]