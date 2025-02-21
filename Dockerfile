FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .

# Instalar as dependências necessárias
RUN apk add --no-cache git

# Baixar dependências
RUN go mod download

# Compilar a aplicação
RUN CGO_ENABLED=0 GOOS=linux go build -o banco-digital ./cmd/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/banco-digital .
COPY weaver.toml .

EXPOSE 8080
CMD ["./banco-digital"] 