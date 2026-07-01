package models

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type Template struct {
	TemplateID  string `json:"template_id"`
	Resolution  Size   `json:"resolution"`
	FPS         int    `json:"fps"`
	Cards       []Card `json:"cards"`
	AudioURL    string `json:"audio_url,omitempty"`
	HWAccel     bool   `json:"hwaccel"`
	JPEGQuality int    `json:"jpeg_quality,omitempty"`
	Subtitles   *bool  `json:"subtitles,omitempty"`
}

type Size struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type VoiceConfig struct {
	Lang  string  `json:"lang,omitempty"`
	Speed float64 `json:"speed,omitempty"`
}

type Card struct {
	ID              string            `json:"id"`
	DurationMs      int               `json:"duration_ms"`
	BackgroundColor string            `json:"background_color"`
	Elements        []Element         `json:"elements"`
	Narration       string            `json:"narration,omitempty"`
	Voice           *VoiceConfig      `json:"voice,omitempty"`
	DurationMode    string            `json:"duration_mode,omitempty"` // "manual", "video", "narration"
	TemplateName    string            `json:"template_name,omitempty"`
	Parameters      map[string]string `json:"parameters,omitempty"`
}

type Element struct {
	Type          string      `json:"type"` // "text", "image", "video", "rect", "circle", "frame", "polygon"
	Content       string      `json:"content,omitempty"`
	FontSize      float64     `json:"font_size,omitempty"`
	Color         string      `json:"color,omitempty"`
	X             float64     `json:"x"`
	Y             float64     `json:"y"`
	Width         float64     `json:"width,omitempty"`
	Height        float64     `json:"height,omitempty"`
	TextAlign     string      `json:"text_align,omitempty"`
	StrokeWidth   float64     `json:"stroke_width,omitempty"`
	Rotation      float64     `json:"rotation,omitempty"`
	Points        [][]float64 `json:"points,omitempty"` // For polygons: [[x1, y1], [x2, y2]]
	ShadowColor   string      `json:"shadow_color,omitempty"`
	ShadowBlur    float64     `json:"shadow_blur,omitempty"`
	ShadowOffsetX float64     `json:"shadow_offset_x,omitempty"`
	ShadowOffsetY float64     `json:"shadow_offset_y,omitempty"`
}

// Validate checks if the template has the required fields
func (t *Template) Validate() error {
	if strings.TrimSpace(t.TemplateID) == "" {
		return errors.New("template_id is required")
	}
	if t.Resolution.Width <= 0 || t.Resolution.Height <= 0 {
		return errors.New("resolution width and height must be greater than 0")
	}
	if t.FPS <= 0 {
		return errors.New("fps must be greater than 0")
	}
	if len(t.Cards) == 0 {
		return errors.New("at least one card is required")
	}

	// 91. Restringir resoluções extremas (max 4K)
	if t.Resolution.Width > 3840 || t.Resolution.Height > 2160 {
		return errors.New("resolução máxima permitida é 4K (3840x2160)")
	}

	// 90. Restringir FPS
	if t.FPS > 60 {
		return errors.New("fps máximo permitido é 60")
	}

	// Validar e atribuir qualidade JPEG padrão
	if t.JPEGQuality == 0 {
		t.JPEGQuality = 2
	}
	if t.JPEGQuality < 1 || t.JPEGQuality > 31 {
		return errors.New("jpeg_quality deve ser entre 1 (melhor) e 31 (pior)")
	}

	for _, card := range t.Cards {
		// 87. Sanitizar ID do card (apenas alfa-numérico ou underscore) para impedir injeção de parâmetros/flags do FFmpeg nos filenames
		if matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, card.ID); !matched {
			return errors.New("card id must be alphanumeric or underscore only to prevent injection: " + card.ID)
		}

		// 90. Restringir duração máxima de cada card (ex: max 30 minutos por card)
		if card.DurationMs <= 0 || card.DurationMs > 1800000 {
			return errors.New("duration_ms must be between 1 and 1800000 ms for card " + card.ID)
		}

		// 92. Validação estrita de cores Hex via Expressão Regular
		if matched, _ := regexp.MatchString(`^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`, card.BackgroundColor); !matched {
			return errors.New("background_color must be a valid hex color for card " + card.ID)
		}
	}

	return nil
}

// GenerateSchemaPrint gera uma representação textual e formatada da estrutura de cards
// e variáveis do template, útil para logs de console e documentações interativas na web.
func (t *Template) GenerateSchemaPrint() string {
	var sb strings.Builder
	sb.WriteString("=========================================================================\n")
	sb.WriteString(fmt.Sprintf(" ESQUEMA DO TEMPLATE: %s\n", t.TemplateID))
	sb.WriteString("=========================================================================\n")
	sb.WriteString(fmt.Sprintf("• Resolução: %dx%d (Aspect Ratio)\n", t.Resolution.Width, t.Resolution.Height))
	sb.WriteString(fmt.Sprintf("• FPS:        %d quadros por segundo\n", t.FPS))
	if t.AudioURL != "" {
		sb.WriteString(fmt.Sprintf("• Trilha Sonora Global: %s\n", t.AudioURL))
	}
	sb.WriteString(fmt.Sprintf("• Aceleração por GPU (NVENC): %t\n", t.HWAccel))
	sb.WriteString(fmt.Sprintf("• Qualidade JPEG Temporário:   %d (escala 1 a 31)\n", t.JPEGQuality))
	sb.WriteString(fmt.Sprintf("• Total de Cenas (Cards):      %d\n", len(t.Cards)))
	sb.WriteString("-------------------------------------------------------------------------\n")

	for i, card := range t.Cards {
		sb.WriteString(fmt.Sprintf(" CENA #%d (ID: %q) | Duração: %.2fs (%d ms) | Fundo: %s\n", 
			i+1, card.ID, float64(card.DurationMs)/1000.0, card.DurationMs, card.BackgroundColor))
		sb.WriteString(" Elementos e variáveis dinâmicas configuráveis:\n")
		
		if len(card.Elements) == 0 {
			sb.WriteString("   (Nenhum elemento nesta cena)\n")
		}
		
		for elIdx, el := range card.Elements {
			badge := ""
			switch el.Type {
			case "video":
				badge = "🎥 VÍDEO"
			case "image":
				badge = "🖼️ IMAGEM"
			case "text":
				badge = "📝 TEXTO"
			case "rect":
				badge = "⏹️ RETÂNGULO"
			case "circle":
				badge = "🟢 CÍRCULO"
			case "polygon":
				badge = "📐 POLÍGONO"
			case "frame":
				badge = "🖼️ MOLDURA"
			default:
				badge = "🧩 ELEMENTO"
			}
			
			// Exibe detalhes específicos baseados no tipo do elemento
			info := ""
			if el.Type == "text" {
				info = fmt.Sprintf("Content: %q, Font Size: %.0f, Color: %s", el.Content, el.FontSize, el.Color)
			} else if el.Type == "video" || el.Type == "image" {
				info = fmt.Sprintf("Content (Path/URL): %q, Size: %.0fx%.0f", el.Content, el.Width, el.Height)
			} else {
				info = fmt.Sprintf("Color: %s, Size: %.0fx%.0f", el.Color, el.Width, el.Height)
			}
			
			sb.WriteString(fmt.Sprintf("   [%d] %-12s | X: %5.0f | Y: %5.0f | %s\n", elIdx+1, badge, el.X, el.Y, info))
		}
		sb.WriteString("-------------------------------------------------------------------------\n")
	}
	
	sb.WriteString(" FORMATO DE ENVIO DO JSON:\n")
	sb.WriteString(" Envie uma requisição POST para /api/render com o corpo correspondente à estrutura acima.\n")
	sb.WriteString("=========================================================================\n")
	return sb.String()
}
