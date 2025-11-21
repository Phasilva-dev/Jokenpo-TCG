FROM golang:1.25.0-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /deployer ./cmd/deployer/main.go

FROM scratch
COPY --from=builder /deployer /deployer
CMD ["/deployer"]