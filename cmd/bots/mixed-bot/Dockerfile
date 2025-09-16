# cmd/bots/mixed-bot/Dockerfile

FROM golang:1.24.3-alpine AS builder
WORKDIR /app

# CORRETO:
COPY go.mod go.sum ./
RUN go mod download

# CORRETO:
COPY . .

# Este caminho já estava correto
RUN CGO_ENABLED=0 GOOS=linux go build -o /bot ./cmd/bots/mixed-bot/main.go

FROM scratch
COPY --from=builder /bot /bot
CMD ["/bot"]