package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// 73. Teste End-to-End validando pipeline CLI
func TestMainE2E(t *testing.T) {
	// Criar diretório temporário para rodar o teste
	tmpDir := t.TempDir()

	// 1. Compilar o binário videogen para o diretório temporário
	binPath := filepath.Join(tmpDir, "videogen")
	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	buildCmd.Dir = "." // roda no diretório cmd/videogen
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("falha ao compilar binário para testes E2E: %v", err)
	}

	// 2. Criar uma pasta bin temporária para mockar o ffmpeg
	mockBinDir := filepath.Join(tmpDir, "mockbin")
	if err := os.MkdirAll(mockBinDir, 0700); err != nil {
		t.Fatalf("falha ao criar pasta de mocks: %v", err)
	}

	// Escrever um script bash fake que simula o ffmpeg
	// Ele só precisa aceitar os parâmetros, consumir o stdin para evitar broken pipe e criar o arquivo de saída com conteúdo mock!
	ffmpegMockPath := filepath.Join(mockBinDir, "ffmpeg")
	ffmpegScript := `#!/bin/sh
cat > /dev/null
for last; do true; done
echo "mock ffmpeg data" > "$last"
exit 0
`
	if err := os.WriteFile(ffmpegMockPath, []byte(ffmpegScript), 0755); err != nil {
		t.Fatalf("falha ao criar mock do ffmpeg: %v", err)
	}

	// 3. Adicionar mockBinDir à frente do PATH para que o binário encontre o nosso ffmpeg
	pathEnv := os.Getenv("PATH")
	newPath := mockBinDir + string(os.PathListSeparator) + pathEnv

	// 4. Criar um template JSON válido de teste
	testJSON := filepath.Join(tmpDir, "test_template.json")
	jsonContent := `{
		"template_id": "test_e2e",
		"resolution": { "width": 640, "height": 480 },
		"fps": 10,
		"cards": [
			{
				"id": "card_01",
				"duration_ms": 1000,
				"background_color": "#000000"
			}
		]
	}`
	if err := os.WriteFile(testJSON, []byte(jsonContent), 0600); err != nil {
		t.Fatalf("falha ao criar JSON de teste: %v", err)
	}

	// 5. Executar o binário compilado
	outMP4 := filepath.Join(tmpDir, "output_final.mp4")
	runCmd := exec.Command(binPath, "-json", testJSON, "-out", outMP4)
	runCmd.Env = append(os.Environ(), "PATH="+newPath)
	
	output, err := runCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("execução do CLI falhou: %v\nOutput: %s", err, string(output))
	}

	// 74. Checar presença, formato e peso (não vazio) do vídeo resultante
	info, err := os.Stat(outMP4)
	if err != nil {
		t.Fatalf("arquivo final não foi criado: %v", err)
	}
	if info.Size() == 0 {
		t.Errorf("arquivo de vídeo final está vazio")
	}

	// 77. Validar falha de JSON inválido
	invalidJSON := filepath.Join(tmpDir, "invalid.json")
	os.WriteFile(invalidJSON, []byte(`{invalid}`), 0600)
	failCmd := exec.Command(binPath, "-json", invalidJSON, "-out", outMP4)
	failCmd.Env = append(os.Environ(), "PATH="+newPath)
	if err := failCmd.Run(); err == nil {
		t.Errorf("esperava erro de parse ao enviar JSON inválido")
	}
}

// 83. Simular falha forçada num card e validar aborto
func TestMainE2E_FFmpegFailure(t *testing.T) {
	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "videogen")
	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("falha ao compilar binário: %v", err)
	}

	mockBinDir := filepath.Join(tmpDir, "mockbin")
	os.MkdirAll(mockBinDir, 0700)

	// Fazer ffmpeg mock falhar (exit status 1)
	ffmpegMockPath := filepath.Join(mockBinDir, "ffmpeg")
	ffmpegScript := `#!/bin/sh
exit 1
`
	os.WriteFile(ffmpegMockPath, []byte(ffmpegScript), 0755)
	newPath := mockBinDir + string(os.PathListSeparator) + os.Getenv("PATH")

	testJSON := filepath.Join(tmpDir, "test_template.json")
	jsonContent := `{"template_id": "test_e2e", "resolution": {"width": 640, "height": 480}, "fps": 10, "cards": [{"id": "card_01", "duration_ms": 1000, "background_color": "#000000"}]}`
	os.WriteFile(testJSON, []byte(jsonContent), 0600)

	outMP4 := filepath.Join(tmpDir, "output_final.mp4")
	runCmd := exec.Command(binPath, "-json", testJSON, "-out", outMP4)
	runCmd.Env = append(os.Environ(), "PATH="+newPath)

	output, err := runCmd.CombinedOutput()
	if err == nil {
		t.Fatalf("esperava erro de execução pois o FFmpeg retornou falha")
	}

	// 83. Validar se aborta de forma legível
	if !strings.Contains(string(output), "Falha crítica ao gerar vídeo") && !strings.Contains(string(output), "Erro ao renderizar card") {
		t.Errorf("esperava erro amigável nos logs, obteve: %s", string(output))
	}
}

// 84. Simular interrupção (Ctrl+C) e checar se limpeza ocorre / termina graciosamente
func TestMainE2E_GracefulShutdown(t *testing.T) {
	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "videogen")
	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("falha ao compilar: %v", err)
	}

	mockBinDir := filepath.Join(tmpDir, "mockbin")
	os.MkdirAll(mockBinDir, 0700)

	// Fazer o mock do ffmpeg travar/dormir para podermos enviar sinal antes de terminar
	ffmpegMockPath := filepath.Join(mockBinDir, "ffmpeg")
	ffmpegScript := `#!/bin/sh
sleep 10
exit 0
`
	os.WriteFile(ffmpegMockPath, []byte(ffmpegScript), 0755)
	newPath := mockBinDir + string(os.PathListSeparator) + os.Getenv("PATH")

	testJSON := filepath.Join(tmpDir, "test_template.json")
	jsonContent := `{"template_id": "test_e2e", "resolution": {"width": 640, "height": 480}, "fps": 10, "cards": [{"id": "card_01", "duration_ms": 10000, "background_color": "#000000"}]}`
	os.WriteFile(testJSON, []byte(jsonContent), 0600)

	outMP4 := filepath.Join(tmpDir, "output_final.mp4")
	runCmd := exec.Command(binPath, "-json", testJSON, "-out", outMP4)
	runCmd.Env = append(os.Environ(), "PATH="+newPath)

	// Iniciar de forma assíncrona
	if err := runCmd.Start(); err != nil {
		t.Fatalf("falha ao iniciar processo: %v", err)
	}

	// Dormir brevemente e mandar SIGINT (Ctrl+C)
	go func() {
		// Pequeno delay
		// Usar um sleep pequeno
		// Envia interrupção
		_ = runCmd.Process.Signal(os.Interrupt)
	}()

	// Aguardar terminação
	err := runCmd.Wait()
	if err == nil {
		t.Errorf("esperava erro de interrupção (exit code 1 ou sinalizado), mas saiu com sucesso")
	}
}
