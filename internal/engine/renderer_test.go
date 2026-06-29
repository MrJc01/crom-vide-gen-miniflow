package engine_test

import (
	"context"
	"image/color"
	"os"
	"path/filepath"
	"testing"
	"videogen/internal/engine"
	"videogen/internal/models"

	"github.com/fogleman/gg"
)

// 82. Testar graceful degradation quando a fonte não existe
func TestRenderer_GracefulDegradation(t *testing.T) {
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "card_out.ts")

	// Criar mock do ffmpeg no PATH para evitar falha do exec.Command
	mockBinDir := filepath.Join(tmpDir, "mockbin")
	if err := os.MkdirAll(mockBinDir, 0700); err != nil {
		t.Fatalf("falha ao criar mockbin: %v", err)
	}
	
	ffmpegMockPath := filepath.Join(mockBinDir, "ffmpeg")
	ffmpegScript := `#!/bin/sh
cat > /dev/null
for last; do true; done
touch "$last"
exit 0
`
	if err := os.WriteFile(ffmpegMockPath, []byte(ffmpegScript), 0755); err != nil {
		t.Fatalf("falha ao mockar ffmpeg: %v", err)
	}

	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", mockBinDir+string(os.PathListSeparator)+oldPath)
	defer os.Setenv("PATH", oldPath)

	card := models.Card{
		ID:              "card_test_degrade",
		DurationMs:      1000,
		BackgroundColor: "#000000",
		Elements: []models.Element{
			{
				Type:     "text",
				Content:  "Hello graceful degradation",
				FontSize: 24,
				Color:    "#FFFFFF",
				X:        100,
				Y:        100,
			},
		},
	}

	renderer := engine.NewFFmpegRenderer(false)
	
	// A fonte assets/fonts/Roboto.ttf não existe no ambiente de teste,
	// então RenderCard vai falhar em dc.LoadFontFace, mas deve continuar executando
	// e gerar o arquivo de saída sem quebras de execução ou pânico!
	err := renderer.RenderCard(context.Background(), card, models.Size{Width: 640, Height: 480}, 10, outPath)
	if err != nil {
		t.Fatalf("RenderCard falhou: %v", err)
	}

	// Verificar se arquivo final foi gerado
	if _, err := os.Stat(outPath); err != nil {
		t.Errorf("arquivo final não foi criado pelo ffmpeg mock: %v", err)
	}
}

// 79. Automatizar testes visuais comparando output isolado vs gabarito de frames
func TestRenderer_VisualComparison(t *testing.T) {
	res := models.Size{Width: 100, Height: 100}
	dc := gg.NewContext(res.Width, res.Height)

	card := models.Card{
		ID:              "card_visual",
		DurationMs:      1000,
		BackgroundColor: "#FF0000", // Fundo vermelho puro
		Elements:        nil,
	}

	engine.DrawCardState(dc, card, res)

	// Gabarito de frames: pixel em (0,0) deve ser vermelho puro
	c := dc.Image().At(0, 0)
	r, g, b, a := c.RGBA()

	// As cores em Go usam multiplicadores de 16 bits (0-65535)
	if r != 0xFFFF || g != 0 || b != 0 || a != 0xFFFF {
		t.Errorf("Esperava pixel vermelho puro (FFFF, 0, 0, FFFF), obteve (%d, %d, %d, %d)", r, g, b, a)
	}

	// 2. Outra cor para garantir
	card.BackgroundColor = "#0000FF" // Azul
	engine.DrawCardState(dc, card, res)
	c = dc.Image().At(0, 0)
	r, g, b, a = c.RGBA()
	if r != 0 || g != 0 || b != 0xFFFF || a != 0xFFFF {
		t.Errorf("Esperava pixel azul puro, obteve (%d, %d, %d, %d)", r, g, b, a)
	}

	// 3. Compara com cor RGB
	expectedColor := color.RGBA{R: 0, G: 0, B: 255, A: 255}
	if r, g, b, a = c.RGBA(); uint8(r>>8) != expectedColor.R || uint8(g>>8) != expectedColor.G || uint8(b>>8) != expectedColor.B || uint8(a>>8) != expectedColor.A {
		t.Errorf("Erro ao mapear cores RGBA correspondentes")
	}
}
