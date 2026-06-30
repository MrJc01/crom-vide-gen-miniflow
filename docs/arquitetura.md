# Arquitetura do Sistema

O **Crom Vide Gen Miniflow** é um gerador de vídeos programático baseado em templates JSON. Ele combina um motor de renderização concorrente em Go com uma interface web moderna em React.

## Diagrama de Componentes

```
┌──────────────────────────────────────────────────────────────────┐
│                        USUÁRIO                                   │
│                                                                  │
│   ┌──────────────┐    ┌──────────────┐    ┌──────────────┐      │
│   │  Navegador   │    │  Terminal    │    │   API REST   │      │
│   │  (React UI)  │    │  (CLI Tool)  │    │  (cURL/etc)  │      │
│   └──────┬───────┘    └──────┬───────┘    └──────┬───────┘      │
└──────────┼───────────────────┼───────────────────┼──────────────┘
           │                   │                   │
           ▼                   │                   │
┌──────────────────┐           │                   │
│   Frontend Web   │           │                   │
│  (React + Vite)  │           │                   │
│   Porta :5173    │           │                   │
└────────┬─────────┘           │                   │
         │ HTTP                │                   │
         ▼                     ▼                   ▼
┌─────────────────────────────────────────────────────────────────┐
│                    API SERVER (Go)                               │
│                      Porta :8080                                │
│                                                                  │
│  ┌─────────────┐  ┌──────────────┐  ┌────────────────────────┐  │
│  │  Templates  │  │  Upload/     │  │  Renderização          │  │
│  │  CRUD       │  │  Media API   │  │  Assíncrona (Goroutine)│  │
│  └──────┬──────┘  └──────┬───────┘  └────────┬───────────────┘  │
│         │                │                   │                   │
│         ▼                ▼                   ▼                   │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                  ENGINE (internal/)                      │    │
│  │                                                         │    │
│  │  ┌──────────┐  ┌───────────┐  ┌──────────────────┐     │    │
│  │  │Renderer  │  │Orchestrat.│  │  DrawCardState    │     │    │
│  │  │(FFmpeg)  │  │(Workers)  │  │  (Frame Drawing)  │     │    │
│  │  └──────────┘  └───────────┘  └──────────────────┘     │    │
│  └─────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
         │                │                   │
         ▼                ▼                   ▼
┌─────────────┐  ┌──────────────┐  ┌──────────────┐
│  templates/ │  │  tmp/uploads │  │   outputs/   │
│  (JSON)     │  │  (Mídias)    │  │   (MP4)      │
└─────────────┘  └──────────────┘  └──────────────┘
```

## Fluxo de Renderização

1. O usuário monta o template no editor web (ou escreve o JSON manualmente).
2. Ao disparar o render, o frontend envia o JSON completo via `POST /api/render`.
3. O servidor cria um **Job** no banco de dados local (`data/videos_db.json`), marca como `"rendering"`, e inicia uma **Goroutine** em background.
4. O **Orchestrator** (`engine/orchestrator.go`) divide os cards entre N workers concorrentes (baseado em `runtime.NumCPU()`).
5. Para cada card, o **Renderer** (`engine/renderer.go`):
   - Extrai frames de vídeos embutidos via `ffmpeg` (para JPEG temporários).
   - Desenha cada frame programaticamente usando a biblioteca `gg` (Go Graphics).
   - Alimenta o `ffmpeg` via pipe para gerar o `.mp4` parcial do card.
6. Após todos os cards serem renderizados, o Orchestrator concatena os `.mp4` parciais em um único arquivo final.
7. O Job é atualizado para `"done"` com o tempo de renderização registrado.

## Camadas do Sistema

| Camada | Pacote Go | Responsabilidade |
|--------|-----------|------------------|
| **Modelos** | `internal/models` | Structs do Template, Card, Element e validação |
| **Motor** | `internal/engine` | Renderização de frames, orquestração e integração FFmpeg |
| **Banco de Dados** | `internal/db` | Persistência de Jobs em arquivo JSON local |
| **Assets Embutidos** | `internal/assets` | Bootstrapping automático de fontes, templates e mídias |
| **Utilitários** | `internal/utils` | Helpers para diretórios temporários e limpeza |
| **Servidor HTTP** | `cmd/server` | API REST, upload de mídias e preview |
| **CLI** | `cmd/videogen` | Ferramenta de linha de comando para renderização headless |
| **Frontend** | `web/` | Interface React + Vite para edição visual |
