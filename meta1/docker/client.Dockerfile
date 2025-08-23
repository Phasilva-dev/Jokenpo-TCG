FROM golang:1.24
#CMD ["sleep", "infinity"]
WORKDIR /app
COPY . .
RUN go build -o server main.go
CMD ["./server"]
