FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY ../../go.mod ../../go.sum ./
RUN go mod download
COPY ../../ .
# Compila o cliente como um binário estático
RUN CGO_ENABLED=0 GOOS=linux go build -o /client ./cmd/client/main.go

FROM scratch
COPY --from=builder /client /client
CMD ["/client"]