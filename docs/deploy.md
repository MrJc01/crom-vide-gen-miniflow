# Guia de Deploy e Produção

Este documento cobre as diferentes formas de colocar o Crom Vide Gen Miniflow em produção.

## 1. Binário Standalone

A forma mais simples de deploy é distribuir o binário compilado do Go. O sistema de `go:embed` embutido garante que todos os assets necessários (fontes, templates padrão) estejam dentro do executável.

### Compilar para a plataforma local

```bash
go build -ldflags="-w -s" -o videogen ./cmd/videogen
go build -ldflags="-w -s" -o videogen-server ./cmd/server
```

### Cross-compilar para múltiplas plataformas

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o dist/videogen-linux-amd64 ./cmd/videogen

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -o dist/videogen-windows-amd64.exe ./cmd/videogen

# macOS Intel
GOOS=darwin GOARCH=amd64 go build -ldflags="-w -s" -o dist/videogen-darwin-amd64 ./cmd/videogen

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -ldflags="-w -s" -o dist/videogen-darwin-arm64 ./cmd/videogen
```

> **Importante:** O FFmpeg deve estar instalado e disponível no `PATH` da máquina de destino.

---

## 2. Docker

O Dockerfile usa multi-stage build para produzir uma imagem compacta baseada em Alpine com FFmpeg.

### Build e execução

```bash
# Build da imagem
docker build -t videogen:latest .

# Executar a CLI
docker run --rm \
  -v $(pwd)/templates:/home/videogen/templates \
  -v $(pwd)/outputs:/home/videogen/output \
  videogen:latest -json=templates/examples/breaking_news.json -out=output/video.mp4
```

### Docker Compose

O `docker-compose.yml` fornecido permite iniciar o servidor API containerizado:

```bash
docker-compose up
```

### Segurança

- A imagem final roda como usuário não-root (`videogen`).
- Apenas os pacotes estritamente necessários são instalados (`ffmpeg`, `ca-certificates`, `tzdata`).
- O binário é compilado com `CGO_ENABLED=0` para portabilidade total.

---

## 3. GitHub Actions (CI/CD)

O projeto inclui dois workflows pré-configurados:

### CI — Testes Automáticos

Arquivo: `.github/workflows/ci.yml`

Roda automaticamente em:
- Push para `main` ou `master`
- Pull Requests para `main` ou `master`

Etapas:
1. Setup do Go 1.22
2. Lint com `golangci-lint`
3. Testes unitários com cobertura
4. Upload do relatório de cobertura para Codecov

### CD — Release Automático

Arquivo: `.github/workflows/cd.yml`

Roda automaticamente ao criar uma **tag** com o padrão `v*` (ex: `v1.0.0`).

Etapas:
1. Setup do Go 1.22
2. Cross-compilação para Linux, Windows e macOS (Intel + ARM)
3. Criação automática de Release no GitHub com os 4 binários anexados

#### Como criar uma release

```bash
# Criar a tag
git tag v1.0.0

# Enviar a tag para o GitHub
git push origin v1.0.0
```

O GitHub Actions irá:
1. Compilar os binários para todas as plataformas
2. Criar a Release automaticamente
3. Anexar os executáveis prontos para download

---

## 4. Variáveis de Ambiente

| Variável | Padrão | Descrição |
|----------|--------|-----------|
| `VIDEOGEN_JSON` | — | Caminho do template padrão para a CLI |
| `VIDEOGEN_OUT` | — | Caminho de saída padrão para a CLI |
| `VIDEOGEN_WORKERS` | — | Número de workers concorrentes padrão |
| `PATH` | — | Deve incluir o diretório do `ffmpeg` e `ffprobe` |

---

## 5. Dependências Externas

| Software | Versão Mínima | Obrigatório | Uso |
|----------|---------------|-------------|-----|
| FFmpeg | 4.x+ | ✅ | Codificação de vídeo, extração de frames |
| FFprobe | 4.x+ | ✅ | Detecção de duração de vídeos enviados |
| Go | 1.21+ | Para compilar | Compilação do backend |
| Node.js | 18+ | Para dev web | Compilação do frontend React |
| Docker | 20+ | Opcional | Deploy containerizado |

---

## 6. Estrutura de Pastas em Produção

```
.
├── videogen-server         # Binário do servidor API
├── videogen                # Binário da CLI (opcional)
├── assets/
│   └── fonts/
│       └── Roboto.ttf      # Extraído automaticamente do binário
├── templates/
│   └── examples/           # Templates JSON (extraídos automaticamente)
├── tmp/
│   └── uploads/            # Mídias enviadas pelos usuários
├── outputs/                # Vídeos MP4 renderizados
├── data/
│   └── videos_db.json      # Banco de dados JSON de jobs
└── bin/
    ├── ffmpeg              # (Opcional) FFmpeg local
    └── ffprobe             # (Opcional) FFprobe local
```

> Se o `ffmpeg` não estiver no `PATH` do sistema, o servidor procura automaticamente na pasta `./bin/` relativa ao diretório de execução.
