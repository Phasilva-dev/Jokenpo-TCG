# --- Estágio 1: Build ---
FROM golang:1.24.3-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# A única mudança é aqui, apontando para o pacote main do cliente.
RUN CGO_ENABLED=0 GOOS=linux go build -o /client ./cmd/client

# --- Estágio 2: Final ---
FROM alpine:latest
COPY --from=builder /client /client

# O cliente será o nosso bot de teste
ENTRYPOINT ["/client"]