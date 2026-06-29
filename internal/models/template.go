package models

import (
	"errors"
	"regexp"
	"strings"
)

type Template struct {
	TemplateID string `json:"template_id"`
	Resolution Size   `json:"resolution"`
	FPS        int    `json:"fps"`
	Cards      []Card `json:"cards"`
	AudioURL   string `json:"audio_url,omitempty"`
}

type Size struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type Card struct {
	ID              string    `json:"id"`
	DurationMs      int       `json:"duration_ms"`
	BackgroundColor string    `json:"background_color"`
	Elements        []Element `json:"elements"`
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

	for _, card := range t.Cards {
		// 87. Sanitizar ID do card (apenas alfa-numérico ou underscore) para impedir injeção de parâmetros/flags do FFmpeg nos filenames
		if matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, card.ID); !matched {
			return errors.New("card id must be alphanumeric or underscore only to prevent injection: " + card.ID)
		}

		// 90. Restringir duração máxima de cada card (ex: max 60s por card)
		if card.DurationMs <= 0 || card.DurationMs > 60000 {
			return errors.New("duration_ms must be between 1 and 60000 ms for card " + card.ID)
		}

		// 92. Validação estrita de cores Hex via Expressão Regular
		if matched, _ := regexp.MatchString(`^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`, card.BackgroundColor); !matched {
			return errors.New("background_color must be a valid hex color for card " + card.ID)
		}
	}

	return nil
}
