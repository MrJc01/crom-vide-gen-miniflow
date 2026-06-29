package utils

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

// 75. Extrair meta-dados via ffprobe
func ProbeVideoMetadata(videoPath string) (string, error) {
	// Verificar se ffprobe está instalado
	if _, err := exec.LookPath("ffprobe"); err != nil {
		return "", fmt.Errorf("ffprobe não encontrado: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// #nosec G204
	cmd := exec.CommandContext(ctx, "ffprobe", 
		"-v", "error", 
		"-show_entries", "format=duration:stream=width,height,r_frame_rate", 
		"-of", "json", 
		videoPath,
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("falha ao rodar ffprobe: %w, output: %s", err, string(out))
	}

	return string(out), nil
}
