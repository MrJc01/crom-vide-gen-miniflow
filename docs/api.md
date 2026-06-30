# Referência da API REST

O servidor HTTP escuta na porta `:8080` e fornece endpoints para gerenciar templates, renderizar vídeos, listar mídias e acompanhar jobs.

## Base URL

```
http://<HOST>:8080
```

> O frontend automaticamente detecta o hostname da máquina via `window.location.hostname`.

---

## Templates

### `GET /api/templates`

Lista todos os templates disponíveis na pasta `templates/examples/`.

**Resposta:** `200 OK`
```json
["breaking_news", "teste", "multiscreen"]
```

---

### `GET /api/templates/{id}`

Retorna o JSON completo de um template específico.

**Exemplo:** `GET /api/templates/breaking_news`

**Resposta:** `200 OK` — O JSON do template.

---

### `POST /api/templates/{id}`

Salva (cria ou sobrescreve) um template no disco.

**Body:** JSON completo do template.

**Resposta:** `200 OK`

---

### `GET /api/templates/{id}/example`

Retorna o JSON do template junto com uma representação textual formatada da sua estrutura (útil para documentação).

**Resposta:** `200 OK`
```json
{
  "template": { ... },
  "schema_print": "=========================================================================\n ESQUEMA DO TEMPLATE: breaking_news\n..."
}
```

---

## Preview

### `POST /api/preview`

Gera uma imagem PNG de preview do primeiro frame do template enviado.

**Body:** JSON do template.

**Resposta:** `200 OK` — Imagem PNG (Content-Type: `image/png`).

---

## Renderização

### `POST /api/render`

Dispara a renderização de um vídeo em background a partir de um template JSON.

**Body:** JSON do template (com `multipart/form-data` se incluir upload de arquivos).

**Resposta:** `200 OK`
```json
{
  "job_id": "job_1782770088459",
  "message": "Rendering started"
}
```

> A renderização é assíncrona. Use os endpoints de vídeos para acompanhar o progresso.

---

## Vídeos (Jobs)

### `GET /api/videos`

Lista todos os jobs de renderização (concluídos, em andamento ou com erro).

**Resposta:** `200 OK`
```json
{
  "job_1782770088459": {
    "id": "job_1782770088459",
    "template_id": "breaking_news",
    "status": "done",
    "file_path": "outputs/output_breaking_news_job_1782770088459.mp4",
    "render_duration_ms": 107633,
    "template": { ... },
    "created_at": "2026-06-29T18:54:48Z",
    "updated_at": "2026-06-29T18:56:36Z"
  }
}
```

---

### `GET /api/videos/{id}`

Retorna os detalhes de um job específico.

---

### `PUT /api/videos/{id}`

Atualiza metadados do job (categoria, arquivamento).

**Body:**
```json
{
  "category": "Marketing",
  "archived": false
}
```

---

### `DELETE /api/videos/{id}`

Exclui o job e o arquivo MP4 associado do disco.

---

## Upload de Mídias

### `POST /api/upload`

Faz upload de um arquivo (vídeo, imagem, áudio) para a pasta `tmp/uploads/`.

**Body:** `multipart/form-data` com campo `file`.

**Resposta:** `200 OK`
```json
{
  "path": "tmp/uploads/meu_video.mp4",
  "duration_ms": 60000
}
```

> O campo `duration_ms` é preenchido automaticamente via `ffprobe` para arquivos de vídeo.

---

## Galeria de Mídias

### `GET /api/media`

Escaneia recursivamente a pasta `tmp/uploads/` e retorna os arquivos organizados por subpastas.

**Resposta:** `200 OK`
```json
{
  "/": [
    {
      "name": "video_intro.mp4",
      "path": "tmp/uploads/video_intro.mp4",
      "url": "http://192.168.18.15:8080/uploads/video_intro.mp4",
      "type": "video",
      "size": 5242880
    }
  ],
  "logos": [
    {
      "name": "logo.png",
      "path": "tmp/uploads/logos/logo.png",
      "url": "http://192.168.18.15:8080/uploads/logos/logo.png",
      "type": "image",
      "size": 24576
    }
  ]
}
```

> As URLs são geradas dinamicamente com base no `Host` da requisição, permitindo acesso via IP da rede local.

---

## Probe de Mídia

### `GET /api/probe?path=...`

Executa `ffprobe` em um arquivo local e retorna a duração.

**Query Params:** `path` — Caminho absoluto ou relativo do arquivo.

**Resposta:** `200 OK`
```json
{
  "duration_ms": 60500
}
```

---

## Arquivos Estáticos

| Rota | Diretório Servido |
|------|-------------------|
| `/outputs/*` | `outputs/` — Vídeos MP4 renderizados |
| `/uploads/*` | `tmp/uploads/` — Mídias enviadas pelo usuário |

---

## CORS

Todas as requisições passam pelo middleware `enableCORS` que permite:
- `Access-Control-Allow-Origin: *`
- `Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS`
- `Access-Control-Allow-Headers: Content-Type`
