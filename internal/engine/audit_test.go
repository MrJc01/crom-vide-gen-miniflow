package engine_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"videogen/internal/engine"
	"videogen/internal/models"
)

// 1. Validação de JSON malformado
func TestAudit_MalformedJSON(t *testing.T) {
	malformedJSON := `{"template_id": "test_tmpl", "resolution": {"width": 1920}, "cards": [` // JSON quebrado
	var tmpl models.Template
	err := json.Unmarshal([]byte(malformedJSON), &tmpl)
	if err == nil {
		t.Errorf("Esperava falha ao fazer parse de JSON malformado, mas obteve sucesso")
	}
}

// 2. Teste de ausência do binário do FFmpeg (o sistema deve falhar graciosamente)
func TestAudit_MissingFFmpeg(t *testing.T) {
	// Temporariamente limpa a variável de ambiente PATH
	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", "")
	defer os.Setenv("PATH", oldPath)

	renderer := engine.NewFFmpegRenderer(false, 2, true)
	card := models.Card{ID: "card_fail", DurationMs: 1000, BackgroundColor: "#000000"}
	res := models.Size{Width: 100, Height: 100}

	err := renderer.RenderCard(context.Background(), card, res, 10, "tmp/out.mp4")
	if err == nil {
		t.Errorf("Esperava erro por ausência do FFmpeg no PATH, mas a execução retornou nil")
	}
}

// 3. Teste de geração de vídeo isolado (garantir que o FFmpeg recebe os raw pixels corretos via pipe)
func TestAudit_IsolatedVideoRawPixels(t *testing.T) {
	tmpDir := t.TempDir()

	mockBinDir := filepath.Join(tmpDir, "mockbin")
	if err := os.MkdirAll(mockBinDir, 0700); err != nil {
		t.Fatalf("falha ao criar mockbin: %v", err)
	}

	logFile := filepath.Join(tmpDir, "ffmpeg_bytes.log")
	ffmpegMock := filepath.Join(mockBinDir, "ffmpeg")
	
	// Script mock do ffmpeg que conta a quantidade exata de bytes recebida no stdin
	ffmpegScript := fmt.Sprintf(`#!/bin/sh
wc -c > "%s"
exit 0
`, logFile)

	if err := os.WriteFile(ffmpegMock, []byte(ffmpegScript), 0755); err != nil {
		t.Fatalf("falha ao mockar ffmpeg: %v", err)
	}

	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", mockBinDir+string(os.PathListSeparator)+oldPath)
	defer os.Setenv("PATH", oldPath)

	renderer := engine.NewFFmpegRenderer(false, 2, true)
	card := models.Card{
		ID:              "card_pixels",
		DurationMs:      1000, // 1 segundo
		BackgroundColor: "#FF0000",
	}
	res := models.Size{Width: 100, Height: 100} // 100x100
	fps := 10 // 10 FPS -> 10 frames no total

	outPath := filepath.Join(tmpDir, "card_out.mp4")
	err := renderer.RenderCard(context.Background(), card, res, fps, outPath)
	if err != nil {
		t.Fatalf("RenderCard falhou com mock: %v", err)
	}

	// Ler arquivo de log com os bytes contados pelo mock
	bytesStr, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("falha ao ler contagem de bytes: %v", err)
	}

	var bytesCount int
	_, err = fmt.Sscanf(string(bytesStr), "%d", &bytesCount)
	if err != nil {
		t.Fatalf("falha ao parsear contagem de bytes: %v", err)
	}

	// Cada frame uncompressed RGBA tem: 100 * 100 * 4 bytes = 40.000 bytes
	// 10 frames a 10 FPS = 400.000 bytes esperados no pipe
	expectedBytes := 10 * 100 * 100 * 4
	if bytesCount != expectedBytes {
		t.Errorf("Tamanho de pixels no pipe incorreto. Esperava %d bytes, obtido %d", expectedBytes, bytesCount)
	}
}

// 4. Teste de timeout/cancelamento de contexto (simulando interrupção pelo usuário)
func TestAudit_TimeoutCancellation(t *testing.T) {
	tmpDir := t.TempDir()

	mockBinDir := filepath.Join(tmpDir, "mockbin")
	if err := os.MkdirAll(mockBinDir, 0700); err != nil {
		t.Fatalf("falha ao criar mockbin: %v", err)
	}

	ffmpegMock := filepath.Join(mockBinDir, "ffmpeg")
	// Mock do ffmpeg que simula uma execução longa
	ffmpegScript := `#!/bin/sh
sleep 15
exit 0
`
	if err := os.WriteFile(ffmpegMock, []byte(ffmpegScript), 0755); err != nil {
		t.Fatalf("falha ao mockar ffmpeg: %v", err)
	}

	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", mockBinDir+string(os.PathListSeparator)+oldPath)
	defer os.Setenv("PATH", oldPath)

	// Contexto cancelado antes do início da renderização
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	renderer := engine.NewFFmpegRenderer(false, 2, true)
	card := models.Card{ID: "card_cancel", DurationMs: 1000, BackgroundColor: "#000000"}
	res := models.Size{Width: 100, Height: 100}

	err := renderer.RenderCard(ctx, card, res, 10, filepath.Join(tmpDir, "cancel.mp4"))
	if err == nil {
		t.Errorf("Esperava falha imediata por contexto cancelado, mas retornou nil")
	}
}
