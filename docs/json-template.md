# Referência do JSON de Template

Os templates são arquivos JSON que descrevem a composição de um vídeo. Este documento explica cada campo, tipo e validação.

## Estrutura Raiz

```json
{
  "template_id": "meu_template",
  "resolution": { "width": 1920, "height": 1080 },
  "fps": 30,
  "audio_url": "caminho/para/trilha.mp3",
  "hwaccel": false,
  "jpeg_quality": 2,
  "cards": [ ... ]
}
```

| Campo | Tipo | Obrigatório | Descrição |
|-------|------|-------------|-----------|
| `template_id` | `string` | ✅ | Identificador único do template (alfanumérico e underscore) |
| `resolution` | `object` | ✅ | Resolução do vídeo em pixels (`width` × `height`) |
| `resolution.width` | `int` | ✅ | Largura em pixels (máx: 3840) |
| `resolution.height` | `int` | ✅ | Altura em pixels (máx: 2160) |
| `fps` | `int` | ✅ | Quadros por segundo (máx: 60) |
| `audio_url` | `string` | ❌ | Caminho relativo ou absoluto para a trilha sonora global |
| `hwaccel` | `bool` | ❌ | Se `true`, tenta usar aceleração GPU (NVENC). Fallback automático para CPU |
| `jpeg_quality` | `int` | ❌ | Qualidade dos JPEGs temporários (1 = melhor, 31 = pior). Padrão: `2` |
| `cards` | `array` | ✅ | Lista de cenas (cards) que compõem o vídeo |

## Objeto Card (Cena)

Cada card representa uma cena do vídeo com duração independente.

```json
{
  "id": "card_intro",
  "duration_ms": 5000,
  "background_color": "#1A1A2E",
  "elements": [ ... ]
}
```

| Campo | Tipo | Obrigatório | Descrição |
|-------|------|-------------|-----------|
| `id` | `string` | ✅ | ID único do card (alfanumérico e underscore) |
| `duration_ms` | `int` | ✅ | Duração da cena em milissegundos (1 a 1.800.000 = 30 min) |
| `background_color` | `string` | ✅ | Cor de fundo em Hex (`#RRGGBB` ou `#RGB`) |
| `elements` | `array` | ❌ | Lista de elementos visuais sobrepostos na cena |

## Objeto Element

Cada elemento é uma camada visual desenhada sobre o card.

### Tipos Suportados

| Tipo | Descrição |
|------|-----------|
| `text` | Texto renderizado com fonte Roboto |
| `image` | Imagem estática (PNG, JPG) sobreposta |
| `video` | Vídeo embutido (extrai frames via FFmpeg) |
| `rect` | Retângulo colorido ou com gradiente |
| `circle` | Círculo colorido |
| `polygon` | Polígono personalizado com pontos definidos |
| `frame` | Moldura decorativa |

### Campos Comuns

| Campo | Tipo | Descrição |
|-------|------|-----------|
| `type` | `string` | Tipo do elemento (ver tabela acima) |
| `x` | `float` | Posição X no canvas |
| `y` | `float` | Posição Y no canvas |
| `width` | `float` | Largura do elemento |
| `height` | `float` | Altura do elemento |
| `color` | `string` | Cor (`#RRGGBB` ou `gradient:#COR1,#COR2`) |
| `rotation` | `float` | Rotação em graus |

### Campos de Texto (`type: "text"`)

| Campo | Tipo | Descrição |
|-------|------|-----------|
| `content` | `string` | O texto a ser exibido |
| `font_size` | `float` | Tamanho da fonte em pixels |
| `text_align` | `string` | Alinhamento: `"left"`, `"center"`, `"right"` |

### Campos de Mídia (`type: "video"` ou `"image"`)

| Campo | Tipo | Descrição |
|-------|------|-----------|
| `content` | `string` | Caminho local ou URL do arquivo de mídia |

### Campos de Sombra (opcionais em qualquer elemento)

| Campo | Tipo | Descrição |
|-------|------|-----------|
| `shadow_color` | `string` | Cor da sombra (`#RRGGBBAA`) |
| `shadow_blur` | `float` | Intensidade do blur |
| `shadow_offset_x` | `float` | Deslocamento horizontal |
| `shadow_offset_y` | `float` | Deslocamento vertical |

### Campos de Polígono (`type: "polygon"`)

| Campo | Tipo | Descrição |
|-------|------|-----------|
| `points` | `array` | Array de pares `[x, y]`: `[[0, 0], [100, 50], [50, 100]]` |

## Gradientes

Para retângulos e formas geométricas, você pode usar gradientes lineares no campo `color`:

```
"color": "gradient:#FF0000,#0000FF"
```

Isso cria um gradiente linear da esquerda (vermelho) para a direita (azul).

## Exemplo Completo Mínimo

```json
{
  "template_id": "hello_world",
  "resolution": { "width": 1280, "height": 720 },
  "fps": 24,
  "cards": [
    {
      "id": "card_1",
      "duration_ms": 3000,
      "background_color": "#0F172A",
      "elements": [
        {
          "type": "text",
          "content": "Hello World!",
          "font_size": 64,
          "color": "#FFFFFF",
          "x": 640,
          "y": 360
        }
      ]
    }
  ]
}
```

## Validações Automáticas

O sistema valida automaticamente ao receber o JSON:

- `template_id` deve ser preenchido.
- Pelo menos 1 card deve existir.
- Resolução máxima: 4K (3840×2160).
- FPS máximo: 60.
- Duração de cada card: 1ms a 1.800.000ms (30 min).
- `background_color` deve ser Hex válido.
- `id` do card deve ser alfanumérico (prevenção de injeção via filenames).
- `jpeg_quality` deve estar entre 1 e 31.
