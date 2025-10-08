# cmd/bots/mixed-bot/Dockerfile

FROM golang:1.25.0-alpine AS builder
WORKDIR /app

# Copia e baixa as dependências primeiro para aproveitar o cache
COPY go.mod go.sum ./
RUN go mod download

# Copia todo o código fonte
COPY . .

# Este caminho já estava correto
RUN CGO_ENABLED=0 GOOS=linux go build -o /bot ./cmd/bots/mixed-bot/main.go

FROM scratch
COPY --from=builder /bot /bot
CMD ["/bot"]