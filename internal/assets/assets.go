package assets

import (
	"embed"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

//go:embed templates
var TemplatesFS embed.FS

//go:embed fonts
var FontsFS embed.FS

//go:embed images
var ImagesFS embed.FS

//go:embed audio
var AudioFS embed.FS

// ExtractAll verifica se os recursos padrões existem no disco rígido local e os extrai
// a partir do binário compilado caso não existam. Isso garante portabilidade total.
func ExtractAll() error {
	// 1. Extrair templates JSON
	if err := extractFS(TemplatesFS, "templates/examples", "templates/examples"); err != nil {
		return err
	}
	// 2. Extrair fonte Roboto.ttf
	if err := extractFS(FontsFS, "fonts", "assets/fonts"); err != nil {
		return err
	}
	// 3. Extrair logo e imagens padrões
	if err := extractFS(ImagesFS, "images", "assets/images"); err != nil {
		return err
	}
	// 4. Extrair arquivos de áudio padrão (MP3) para a raiz do projeto
	if err := extractFS(AudioFS, "audio", "."); err != nil {
		return err
	}
	return nil
}

// extractFS é um helper recursivo para ler o sistema de arquivos virtual embutido no binário
// e gravar fisicamente no disco rígido local, sem sobrescrever arquivos já existentes.
func extractFS(fs embed.FS, srcDir, destDir string) error {
	entries, err := fs.ReadDir(srcDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(srcDir, entry.Name())
		destPath := filepath.Join(destDir, entry.Name())

		if entry.IsDir() {
			err = extractFS(fs, srcPath, destPath)
			if err != nil {
				return err
			}
			continue
		}

		// Importante: Não sobrescreve modificações locais feitas pelo usuário
		if _, err := os.Stat(destPath); err == nil {
			continue
		}

		// Garante a existência do diretório pai
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		// Abre o arquivo embutido
		srcFile, err := fs.Open(srcPath)
		if err != nil {
			return err
		}

		// Grava o arquivo físico
		destFile, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			srcFile.Close()
			return err
		}

		_, err = io.Copy(destFile, srcFile)
		srcFile.Close()
		destFile.Close()
		if err != nil {
			return err
		}

		slog.Info("Recurso embutido bootstrap extraído com sucesso", "origem", srcPath, "destino", destPath)
	}

	return nil
}
