package models_test

import (
	"encoding/json"
	"testing"
	"videogen/internal/models"
)

func TestTemplateValidation_Success(t *testing.T) {
	validJSON := `{
		"template_id": "promo_v1",
		"resolution": { "width": 1080, "height": 1920 },
		"fps": 30,
		"cards": [
			{
				"id": "card_01",
				"duration_ms": 3000,
				"background_color": "#1A1A1A"
			}
		]
	}`

	var tmpl models.Template
	if err := json.Unmarshal([]byte(validJSON), &tmpl); err != nil {
		t.Fatalf("falha no unmarshal: %v", err)
	}

	if err := tmpl.Validate(); err != nil {
		t.Errorf("esperava sucesso, mas deu erro: %v", err)
	}
}

func TestTemplateValidation_Failures(t *testing.T) {
	tests := []struct {
		name        string
		jsonContent string
		expectError string
	}{
		{
			name: "Missing Template ID",
			jsonContent: `{"resolution": {"width": 1080, "height": 1920}, "fps": 30, "cards": [{"id": "c1", "duration_ms": 1000, "background_color": "#FFFFFF"}]}`,
			expectError: "template_id is required",
		},
		{
			name: "Invalid Resolution (Width 0)",
			jsonContent: `{"template_id": "t1", "resolution": {"width": 0, "height": 1920}, "fps": 30, "cards": [{"id": "c1", "duration_ms": 1000, "background_color": "#FFFFFF"}]}`,
			expectError: "resolution width and height must be greater than 0",
		},
		{
			name: "Resolution Too Large (5K)",
			jsonContent: `{"template_id": "t1", "resolution": {"width": 5120, "height": 2880}, "fps": 30, "cards": [{"id": "c1", "duration_ms": 1000, "background_color": "#FFFFFF"}]}`,
			expectError: "resolução máxima permitida é 4K (3840x2160)",
		},
		{
			name: "FPS Too Large (120)",
			jsonContent: `{"template_id": "t1", "resolution": {"width": 1920, "height": 1080}, "fps": 120, "cards": [{"id": "c1", "duration_ms": 1000, "background_color": "#FFFFFF"}]}`,
			expectError: "fps máximo permitido é 60",
		},
		{
			name: "Missing Cards",
			jsonContent: `{"template_id": "t1", "resolution": {"width": 1080, "height": 1920}, "fps": 30, "cards": []}`,
			expectError: "at least one card is required",
		},
		{
			name: "Invalid Card Color",
			jsonContent: `{"template_id": "t1", "resolution": {"width": 1080, "height": 1920}, "fps": 30, "cards": [{"id": "c1", "duration_ms": 1000, "background_color": "FFFFFF"}]}`,
			expectError: "background_color must be a valid hex color for card c1",
		},
		{
			name: "Card ID Injection Attempt",
			jsonContent: `{"template_id": "t1", "resolution": {"width": 1920, "height": 1080}, "fps": 30, "cards": [{"id": "c1;rm -rf /", "duration_ms": 1000, "background_color": "#FFFFFF"}]}`,
			expectError: "card id must be alphanumeric or underscore only to prevent injection: c1;rm -rf /",
		},
		{
			name: "Card Duration Too Long (30m)",
			jsonContent: `{"template_id": "t1", "resolution": {"width": 1920, "height": 1080}, "fps": 30, "cards": [{"id": "c1", "duration_ms": 1800001, "background_color": "#FFFFFF"}]}`,
			expectError: "duration_ms must be between 1 and 1800000 ms for card c1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tmpl models.Template
			err := json.Unmarshal([]byte(tt.jsonContent), &tmpl)
			if err != nil {
				t.Fatalf("falha no parse json mock: %v", err)
			}
			err = tmpl.Validate()
			if err == nil {
				t.Errorf("esperava erro, mas não deu erro")
			} else if err.Error() != tt.expectError {
				t.Errorf("erro esperado: %q, recebido: %q", tt.expectError, err.Error())
			}
		})
	}
}
