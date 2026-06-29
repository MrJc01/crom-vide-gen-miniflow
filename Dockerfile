# 56. Multi-stage Dockerfile
# Stage 1: Build
FROM golang:1.22-alpine AS builder

# Instalar dependências necessárias para build (se houver, cgo etc)
# Para puramente go, alpine puro basta.
RUN apk update && apk add --no-cache git

WORKDIR /app

# 61. Cache de dependências (copiar go.mod primeiro)
COPY go.mod go.sum ./
RUN go mod download

# Copiar resto do código
COPY . .

# Build do binário estático
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o videogen ./cmd/videogen


# Stage 2: Produção (Run)
# 57. Incluir FFmpeg na imagem final
FROM alpine:3.19

# 65. Otimizar tamanho (alpine + apenas pacotes estritamente necessários)
RUN apk add --no-cache ffmpeg ca-certificates tzdata

# 64. Restrições de Segurança: não rodar como root
RUN addgroup -S videogen && adduser -S videogen -G videogen
USER videogen

WORKDIR /home/videogen

# Copiar binário do builder
COPY --from=builder /app/videogen .

# Diretório para montar os templates (read) e a pasta temporária/output (write)
RUN mkdir -p tmp output templates assets

# Entrypoint do sistema
ENTRYPOINT ["./videogen"]
CMD ["-help"]
