package utils_test

import (
	"os"
	"path/filepath"
	"testing"
	"videogen/internal/utils"
)

func TestEnsureDirectories(t *testing.T) {
	tmpDir := filepath.Join(t.TempDir(), "sub/dir")
	err := utils.EnsureDirectories(tmpDir)
	if err != nil {
		t.Fatalf("EnsureDirectories falhou: %v", err)
	}

	info, err := os.Stat(tmpDir)
	if err != nil {
		t.Fatalf("Diretório não foi criado: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("Caminho não é um diretório")
	}
}

func TestSanitizePath(t *testing.T) {
	basePath := "/home/user/app"
	
	tests := []struct {
		name      string
		userPath  string
		shouldErr bool
	}{
		{"Safe Relative Path", "/home/user/app/templates/v1.json", false},
		{"Directory Traversal Attack", "/home/user/app/../other/app", true},
		{"Root traversal", "/etc/passwd", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := utils.SanitizePath(basePath, tt.userPath)
			if (err != nil) != tt.shouldErr {
				t.Errorf("SanitizePath(%q) erro esperado: %t, erro obtido: %v", tt.userPath, tt.shouldErr, err)
			}
		})
	}
}

func TestSafeDiv(t *testing.T) {
	if val := utils.SafeDiv(10, 2); val != 5 {
		t.Errorf("Esperava 5, obtido %d", val)
	}
	if val := utils.SafeDiv(10, 0); val != 0 {
		t.Errorf("Divisão por zero esperava 0, obtido %d", val)
	}
}

func TestFileExists(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "test.txt")
	if utils.FileExists(tempFile) {
		t.Errorf("Caminho não deveria existir")
	}

	if err := os.WriteFile(tempFile, []byte("ok"), 0644); err != nil {
		t.Fatalf("Erro ao criar arquivo de teste: %v", err)
	}

	if !utils.FileExists(tempFile) {
		t.Errorf("Caminho deveria existir")
	}
}

func TestValidateFileSize(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "size.txt")
	content := []byte("hello world") // 11 bytes
	if err := os.WriteFile(tempFile, content, 0644); err != nil {
		t.Fatalf("falha ao escrever: %v", err)
	}

	// 1. Caso de sucesso
	if err := utils.ValidateFileSize(tempFile, 20); err != nil {
		t.Errorf("não deveria falhar para limite 20: %v", err)
	}

	// 2. Caso de falha (excede tamanho)
	if err := utils.ValidateFileSize(tempFile, 5); err == nil {
		t.Errorf("deveria falhar para limite 5")
	}
}

func TestValidateMIMEType(t *testing.T) {
	// PNG válido (MIME detectado via magic numbers)
	// Primeiro byte de PNG: 137 80 78 71 13 10 26 10
	pngHeader := []byte{137, 80, 78, 71, 13, 10, 26, 10, 0, 0, 0, 0}
	tempPNG := filepath.Join(t.TempDir(), "img.png")
	if err := os.WriteFile(tempPNG, pngHeader, 0644); err != nil {
		t.Fatalf("falha ao escrever png mock: %v", err)
	}

	if err := utils.ValidateMIMEType(tempPNG); err != nil {
		t.Errorf("deveria permitir PNG válido: %v", err)
	}

	// TXT / Binário não permitido
	tempTXT := filepath.Join(t.TempDir(), "invalid.txt")
	if err := os.WriteFile(tempTXT, []byte("hello format plain text"), 0644); err != nil {
		t.Fatalf("falha ao escrever txt mock: %v", err)
	}
	if err := utils.ValidateMIMEType(tempTXT); err == nil {
		t.Errorf("deveria rejeitar texto plano (MIME text/plain não permitido)")
	}

	// Arquivo inexistente
	if err := utils.ValidateMIMEType("non_existent_file.png"); err == nil {
		t.Errorf("deveria falhar para arquivo inexistente")
	}
}

func TestCleanupTempFiles(t *testing.T) {
	tempDir := t.TempDir()
	
	f1 := filepath.Join(tempDir, "file1.ts")
	f2 := filepath.Join(tempDir, "file2.ts")
	
	os.WriteFile(f1, []byte("ts1"), 0644)
	os.WriteFile(f2, []byte("ts2"), 0644)

	if err := utils.CleanupTempFiles(tempDir); err != nil {
		t.Fatalf("CleanupTempFiles falhou: %v", err)
	}

	entries, _ := os.ReadDir(tempDir)
	if len(entries) != 0 {
		t.Errorf("Diretório temp não foi limpo completamente")
	}
}
