package utils

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// 11. Gerenciador de arquivos temporários e diretório de saída
func EnsureDirectories(dirs ...string) error {
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0750); err != nil {
			return fmt.Errorf("falha ao criar diretório %s: %w", dir, err)
		}
	}
	return nil
}

// 63. Limpeza pós-execução
func CleanupTempFiles(tmpDir string) error {
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("erro ao ler diretório temp: %w", err)
	}

	for _, entry := range entries {
		path := filepath.Join(tmpDir, entry.Name())
		if err := os.Remove(path); err != nil {
			slog.Warn("Falha ao remover arquivo temporário", "path", path, "erro", err)
		}
	}
	slog.Info("Limpeza de temporários concluída", "dir", tmpDir)
	return nil
}

// 88. Sanitizar paths contra Directory Traversal
func SanitizePath(basePath, userPath string) (string, error) {
	absBase, err := filepath.Abs(basePath)
	if err != nil {
		return "", err
	}

	absUser, err := filepath.Abs(userPath)
	if err != nil {
		return "", err
	}

	if !strings.HasPrefix(absUser, absBase) {
		return "", fmt.Errorf("path traversal detectado: %s está fora de %s", userPath, basePath)
	}

	return absUser, nil
}

// 89. Payload limit
const MaxJSONSize = 10 * 1024 * 1024 // 10MB

func ValidateFileSize(path string, maxBytes int64) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("erro ao checar arquivo %s: %w", path, err)
	}
	if info.Size() > maxBytes {
		return fmt.Errorf("arquivo %s excede o limite de %d bytes (tamanho: %d)", path, maxBytes, info.Size())
	}
	return nil
}

// 98. MIME whitelist
var AllowedMIMETypes = map[string]bool{
	"image/png":              true,
	"image/jpeg":             true,
	"image/gif":              true,
	"image/webp":             true,
	"font/ttf":               true,
	"font/otf":               true,
	"application/x-font-ttf": true,
}

func ValidateMIMEType(path string) error {
	// #nosec G304
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	buffer := make([]byte, 512)
	n, err := f.Read(buffer)
	if err != nil {
		return err
	}

	mimeType := http.DetectContentType(buffer[:n])
	// Extrair base MIME (sem parametros charset etc)
	baseMIME := strings.Split(mimeType, ";")[0]
	baseMIME = strings.TrimSpace(baseMIME)

	if !AllowedMIMETypes[baseMIME] {
		return fmt.Errorf("tipo MIME não permitido: %s (arquivo: %s)", baseMIME, path)
	}

	return nil
}

// 97. Prevenção de divisão por zero
func SafeDiv(a, b int) int {
	if b == 0 {
		return 0
	}
	return a / b
}

// 96. Proteção contra sobrescrita
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
