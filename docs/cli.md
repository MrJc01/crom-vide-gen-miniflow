# CLI `videogen` — Referência de Uso

A ferramenta de linha de comando `videogen` permite renderizar vídeos a partir de templates JSON sem a interface web.

## Uso Básico

```bash
videogen -json=template.json -out=video.mp4
```

## Flags Disponíveis

| Flag | Tipo | Padrão | Descrição |
|------|------|--------|-----------|
| `-json` | `string` | `templates/sample_v1.json` | Caminho do arquivo JSON de template |
| `-out` | `string` | `output_final.mp4` | Caminho e nome do arquivo de saída MP4 |
| `-workers` | `int` | `0` (auto) | Número de workers concorrentes. `0` = usa todos os núcleos da CPU |
| `-hwaccel` | `bool` | `false` | Habilita aceleração por hardware (NVENC) |
| `-schema` | `bool` | `false` | Exibe o esquema textual formatado do template e encerra |
| `-version` | `bool` | `false` | Exibe a versão do sistema e encerra |
| `-cpuprofile` | `string` | `""` | Grava o profile de CPU no arquivo especificado (para diagnóstico) |
| `-memprofile` | `string` | `""` | Grava o profile de memória no arquivo especificado |

## Variáveis de Ambiente

| Variável | Descrição | Prioridade |
|----------|-----------|------------|
| `VIDEOGEN_JSON` | Caminho padrão do template | Inferior à flag `-json` |
| `VIDEOGEN_OUT` | Caminho de saída padrão | Inferior à flag `-out` |
| `VIDEOGEN_WORKERS` | Número de workers padrão | Inferior à flag `-workers` |

## Exemplos

### Renderizar um vídeo

```bash
videogen -json=templates/examples/breaking_news.json -out=breaking.mp4
```

### Visualizar o esquema de um template

```bash
videogen -json=templates/examples/breaking_news.json -schema
```

Saída:
```
=========================================================================
 ESQUEMA DO TEMPLATE: breaking_news
=========================================================================
• Resolução: 1920x1080 (Aspect Ratio)
• FPS:        30 quadros por segundo
• Trilha Sonora Global: the_mountain-piano-background-487020.mp3
• Aceleração por GPU (NVENC): false
• Qualidade JPEG Temporário:   2 (escala 1 a 31)
• Total de Cenas (Cards):      1
-------------------------------------------------------------------------
 CENA #1 (ID: "card_news") | Duração: 6.00s (6000 ms) | Fundo: #00FF00
 Elementos e variáveis dinâmicas configuráveis:
   [1] 🎥 VÍDEO      | X:   960 | Y:   540 | Content (Path/URL): "video.mp4", Size: 1920x1080
   [2] ⏹️ RETÂNGULO  | X:   250 | Y:    80 | Color: gradient:#CC0000,#880000, Size: 500x80
   [3] 📝 TEXTO      | X:   250 | Y:    80 | Content: "BREAKING NEWS", Font Size: 40, Color: #FFFFFF
=========================================================================
```

### Renderizar com aceleração GPU e profiling

```bash
videogen -json=template.json -out=video.mp4 -hwaccel -cpuprofile=cpu.prof
```

### Usar com 4 workers explícitos

```bash
videogen -json=template.json -out=video.mp4 -workers=4
```

## Código de Saída

| Código | Significado |
|--------|-------------|
| `0` | Sucesso |
| `1` | Erro (JSON inválido, FFmpeg não encontrado, falha na renderização) |

## Bootstrapping Automático

Ao iniciar, o binário verifica se as pastas de assets (`assets/fonts/`, `templates/examples/`, etc.) existem no disco. Se não existirem, os recursos embutidos no binário via `go:embed` são extraídos automaticamente. Isso permite distribuir o binário como um único executável portátil.
