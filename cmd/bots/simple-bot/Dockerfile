# cmd/bots/simple-login/Dockerfile

FROM golang:1.24.3-alpine AS builder
WORKDIR /app

# ANTES: COPY ../../../go.mod ../../../go.sum ./
# CORRETO: Copia a partir da raiz do projeto (o contexto de build)
COPY go.mod go.sum ./
RUN go mod download

# ANTES: COPY ../../../ .
# CORRETO: Copia a partir da raiz do projeto
COPY . .

# Este caminho já estava correto, pois é executado dentro do container
RUN CGO_ENABLED=0 GOOS=linux go build -o /bot ./cmd/bots/simple-login/main.go

FROM scratch
COPY --from=builder /bot /bot
CMD ["/bot"]