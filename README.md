# Crom Vide Gen Miniflow 🎬

Motor de geração programática de vídeos baseado em templates JSON, com renderização concorrente em Go, interface web em React e deploy cross-platform automatizado.

[![Go CI](https://github.com/MrJc01/crom-vide-gen-miniflow/actions/workflows/ci.yml/badge.svg)](https://github.com/MrJc01/crom-vide-gen-miniflow/actions/workflows/ci.yml)

---

## ✨ Funcionalidades

- **Templates JSON declarativos** — Defina vídeos inteiros com textos, retângulos, círculos, polígonos, imagens e vídeos embutidos.
- **Renderização concorrente** — Cards (cenas) são processados em paralelo por workers do Go.
- **Interface web completa** — Editor visual, preview em tempo real, timeline com drag/clone e gerenciador de mídias.
- **Galeria de Mídias** — Navegação por pastas, upload, preview e seleção integrada ao editor.
- **Aceleração GPU** — Suporte a NVENC (NVIDIA) com fallback automático para CPU.
- **Otimização de I/O** — Cache de fontes e imagens em memória + controle de qualidade JPEG.
- **Trilha sonora global** — Áudio MP3 mixado automaticamente no vídeo final.
- **Gradientes** — Suporte a gradientes lineares em retângulos e formas.
- **Schema textual** — Impressão formatada do template via CLI ou API para documentação.
- **Bootstrapping automático** — Assets embutidos no binário via `go:embed`, extraídos ao primeiro uso.
- **Cross-platform** — Binários para Linux, macOS (Intel + ARM) e Windows via GitHub Actions.
- **Docker** — Multi-stage Dockerfile otimizado com Alpine + FFmpeg.

---

## 🚀 Início Rápido

### Pré-requisitos

- **Go 1.21+**
- **Node.js 18+ e npm**
- **FFmpeg** e **FFprobe** disponíveis no `PATH`
- (Opcional) Docker e Docker Compose

### 1. Iniciar a API (Backend)

```bash
go run cmd/server/main.go
```

> O servidor inicia na porta `:8080`.

### 2. Iniciar o Frontend (React)

```bash
cd web
npm install
npm run dev -- --host
```

> Acesse `http://localhost:5173` ou pelo IP da rede local.

### 3. Renderizar via CLI (sem interface web)

```bash
go run cmd/videogen/main.go -json=templates/examples/breaking_news.json -out=video.mp4
```

### 4. Visualizar o esquema de um template

```bash
go run cmd/videogen/main.go -json=templates/examples/breaking_news.json -schema
```

### 5. Via Docker

```bash
make docker-run
```

---

## 📁 Estrutura do Projeto

```
.
├── cmd/
│   ├── server/         # Servidor HTTP (API REST)
│   └── videogen/       # Ferramenta de linha de comando
├── internal/
│   ├── assets/         # Assets embutidos (go:embed)
│   ├── db/             # Banco de dados JSON local (jobs)
│   ├── engine/         # Motor de renderização (FFmpeg + desenho)
│   ├── models/         # Structs de Template, Card, Element
│   └── utils/          # Helpers de diretórios e limpeza
├── web/                # Frontend React + Vite
├── templates/examples/ # Templates JSON de exemplo
├── assets/fonts/       # Fonte Roboto (extraída automaticamente)
├── docs/               # Documentação detalhada
├── .github/workflows/  # CI (testes) e CD (release)
├── Dockerfile          # Multi-stage build otimizado
├── Makefile            # Automações de build, test e deploy
└── docker-compose.yml  # Orquestração Docker
```

---

## 📚 Documentação

| Documento | Descrição |
|-----------|-----------|
| [Arquitetura](docs/arquitetura.md) | Diagrama de componentes, fluxo de renderização e camadas do sistema |
| [JSON Template](docs/json-template.md) | Referência completa dos campos, tipos e validações do JSON de template |
| [API REST](docs/api.md) | Documentação de todos os endpoints HTTP com exemplos |
| [CLI](docs/cli.md) | Flags, variáveis de ambiente e exemplos de uso da ferramenta de linha de comando |
| [Interface Web](docs/interface-web.md) | Guia do usuário para o editor, galeria de mídias e player de vídeos |
| [Deploy](docs/deploy.md) | Cross-compilação, Docker, GitHub Actions (CI/CD) e deploy em produção |

---

## 🔧 Comandos do Makefile

```bash
make build        # Compila o binário videogen
make run          # Compila e executa com o template padrão
make test         # Roda testes unitários com cobertura
make clean        # Limpa artefatos temporários
make docker-build # Build da imagem Docker
make docker-run   # Build + execução via Docker
make lint         # Lint com golangci-lint
```

---

## 🏷️ Releases

Releases são criadas automaticamente pelo GitHub Actions ao criar uma tag:

```bash
git tag v1.0.0
git push origin v1.0.0
```

Binários disponíveis em cada release:

| Plataforma | Arquivo |
|------------|---------|
| Linux (AMD64) | `videogen-linux-amd64` |
| Windows (AMD64) | `videogen-windows-amd64.exe` |
| macOS (Intel) | `videogen-darwin-amd64` |
| macOS (Apple Silicon) | `videogen-darwin-arm64` |

---

## 🧪 Testes

```bash
# Rodar todos os testes
go test ./...

# Com cobertura e verbose
go test -v -cover ./...
```

---

## 📄 Licença

Este projeto é distribuído sob os termos da licença do repositório.
