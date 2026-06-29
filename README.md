# Crom Vide Gen Miniflow 🎬

Este projeto é um gerador de vídeos baseado em templates JSON, com foco em simplicidade, concorrência e escalabilidade. Ele possui uma ferramenta de linha de comando (CLI) ultra-rápida construída em Go e uma interface web (React) para visualizar e editar os templates, com o apoio de uma API.

## 🏗 Arquitetura do Projeto

O sistema é dividido em três componentes principais:

1. **CLI `videogen` (Back-end Core)**: Motor escrito em Go que processa os templates JSON, renderiza os frames utilizando processamento concorrente (workers) e orquestra a geração do arquivo final MP4 via **FFmpeg**.
2. **API Server `server` (Back-end API)**: Um servidor HTTP (Go) que permite a interface web listar templates, salvá-los e gerar um preview (PNG) na hora do primeiro frame de um card.
3. **Web UI `web` (Front-end)**: Aplicação em React + Vite para gerenciar os templates JSON e ver os previews renderizados.

---

## 🚀 Como Iniciar

### Pré-requisitos
- **Go 1.21+** instalado.
- **Node.js + npm** instalados (para rodar o web UI).
- **FFmpeg** instalado na máquina e disponível no `PATH`.
- (Opcional) **Docker e Docker Compose** caso queira rodar conteinerizado.

---

### 1. Rodando a Interface de Criação (Web + Server API)

Para criar ou modificar seus vídeos de maneira visual, você precisará iniciar o servidor da API em Go e o frontend em React.

**Passo A: Iniciar a API (Backend)**
Abra o terminal na raiz do projeto e rode:
```bash
go run cmd/server/main.go
```
> O servidor iniciará na porta `8080` (http://localhost:8080) e ficará pronto para processar templates.

**Passo B: Iniciar a Interface (Frontend)**
Em um novo terminal, entre na pasta `web` e rode:
```bash
cd web
npm install
npm run dev
```
> Acesse o endereço indicado no terminal (geralmente `http://localhost:5173`). Lá você poderá editar os arquivos JSON da pasta `templates/examples` e ver o preview na tela!

---

### 2. Gerando o Vídeo (CLI `videogen`)

Quando seu JSON estiver pronto e você quiser renderizar o vídeo final, use a CLI em Go.

Você pode usar o **Makefile** para simplificar:

```bash
# Para compilar o binário videogen e já rodar com o template de exemplo:
make run
```

Se quiser passar um template específico:
```bash
# Compile o binário
make build

# Execute passando os parâmetros:
./videogen -json=templates/examples/SEU_TEMPLATE.json -out=meu_video.mp4
```

**Parâmetros suportados pela CLI:**
- `-json`: Caminho para o JSON do template (padrão: `templates/sample_v1.json`).
- `-out`: Caminho de destino para o vídeo `.mp4` (padrão: `output_final.mp4`).
- `-workers`: Número de rotinas simultâneas (o padrão é utilizar todos os núcleos da CPU).
- `-hwaccel`: Habilita aceleração de hardware do FFmpeg se estiver disponível.

---

### 3. Rodando via Docker (Opcional)

Você pode compilar e rodar a geração do vídeo via Docker sem precisar instalar Go ou FFmpeg no seu computador principal.

```bash
# Isso irá mapear as pastas ./templates e ./output, gerar a imagem e rodar o videogen
make docker-run
```
O vídeo final irá aparecer na pasta `output/`.

---

## 📁 Estrutura de Diretórios

- `/cmd/videogen`: Código principal do executável de geração de vídeos.
- `/cmd/server`: API Rest para lidar com os pedidos da UI.
- `/internal`: Lógica de negócios (engine de desenho, controle do FFmpeg).
- `/templates`: Arquivos JSON de exemplo e assets utilizados pelos templates.
- `/web`: Interface em React (Vite).
- `Makefile`: Automações de compilação, teste e docker.
- `docker-compose.yml`: Definição de recursos e serviço para rodar isolado no Docker.
