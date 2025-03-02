FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .

# Instalar as dependências necessárias
RUN apk add --no-cache git

# Baixar dependências
RUN go mod download
RUN go mod tidy

# Compilar a aplicação
RUN CGO_ENABLED=0 GOOS=linux go build -o banco-digital ./cmd/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/banco-digital .

# Configurar variáveis de ambiente para logs não bufferizados
ENV GOTRACEBACK=single \
    GOGC=off \
    GOMEMLIMIT=1000MiB \
    GODEBUG=gctrace=0 \
    GOMAXPROCS=1

EXPOSE 8080
CMD ["./banco-digital"]