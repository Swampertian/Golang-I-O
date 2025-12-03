# ===============================
# STAGE 1 — Build da aplicação
# ===============================
FROM golang:1.23-alpine AS builder


# Instalar build deps opcionais
RUN apk add --no-cache git

WORKDIR /app

# Copia mod files para cache rápido
COPY go.mod go.sum ./
RUN go mod download

# Copia todo o restante
COPY . .

# Build final estático (CGO disabled)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o fire-go .

# ===============================
# STAGE 2 — Imagem minimalista
# ===============================
FROM alpine:3.19

WORKDIR /app

# Copiar binário do build stage
COPY --from=builder /app/fire-go .

# Porta da API
EXPOSE 8081

# Healthcheck opcional
HEALTHCHECK --interval=10s --timeout=3s \
    CMD wget -qO- http://localhost:8081/health || exit 1

CMD ["./fire-go"]
