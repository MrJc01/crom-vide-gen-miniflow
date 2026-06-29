package engine

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
	"videogen/internal/models"
	"videogen/internal/utils"
)

// Worker Pool Pattern
func renderWorker(ctx context.Context, id int, jobs <-chan models.Card, results chan<- string, tmpl models.Template, wg *sync.WaitGroup, renderer Renderer) {
	defer wg.Done()
	for card := range jobs {
		// 52. Prevenir vazamento de goroutines no caso de interrupções
		select {
		case <-ctx.Done():
			slog.Warn("Cancelando render worker", "worker", id)
			return
		default:
		}

		outPath := filepath.Join("tmp", fmt.Sprintf("%s.ts", card.ID))
		slog.Info("Renderizando card", "worker", id, "card", card.ID, "outPath", outPath)

		// 48. Coletar métricas de tempo de execução por card
		start := time.Now()
		err := renderer.RenderCard(card, tmpl.Resolution, tmpl.FPS, outPath)
		elapsed := time.Since(start)

		if err != nil {
			slog.Error("Erro ao renderizar card", "card", card.ID, "erro", err, "duracao_ms", elapsed.Milliseconds())
			results <- "" // emite vazio em caso de erro (poderia passar struct de erro)
			continue
		}
		
		slog.Info("Card renderizado com sucesso", "card", card.ID, "duracao_ms", elapsed.Milliseconds())
		results <- outPath
	}
}

// 8. ProcessVideo Orquestra as chamadas
func ProcessVideo(ctx context.Context, tmpl models.Template, finalOutput string, numWorkers int, renderer Renderer) error {
	// 41. sync.WaitGroup e 42. Worker Pool
	var wg sync.WaitGroup

	jobs := make(chan models.Card, len(tmpl.Cards))
	results := make(chan string, len(tmpl.Cards))

	// Inicia os workers
	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go renderWorker(ctx, w, jobs, results, tmpl, &wg, renderer)
	}

	// Envia os jobs
	for _, card := range tmpl.Cards {
		jobs <- card
	}
	close(jobs) // Nenhum outro job será enviado

	wg.Wait()
	close(results)

	var tmpFiles []string
	for res := range results {
		if res != "" {
			tmpFiles = append(tmpFiles, res)
		}
	}

	if len(tmpFiles) != len(tmpl.Cards) {
		return fmt.Errorf("falha ao renderizar todos os cards")
	}

	var audioPath string
	if tmpl.AudioURL != "" {
		slog.Info("Trilha sonora especificada. Baixando áudio...", "url", tmpl.AudioURL)
		var err error
		audioPath, err = utils.DownloadRemoteFile(tmpl.AudioURL, "tmp")
		if err != nil {
			return fmt.Errorf("falha ao baixar áudio remoto: %w", err)
		}
	}

	slog.Info("Todos os cards renderizados! Juntando partes...")
	return ConcatVideos(tmpFiles, audioPath, finalOutput)
}

// 34 e 35. Demux/Concat
func ConcatVideos(files []string, audioPath string, finalOutput string) error {
	listPath := "tmp/list.txt"

	// 51. Otimizar I/O (bufferizado)
	f, err := os.Create(listPath)
	if err != nil {
		return fmt.Errorf("erro ao criar arquivo list.txt: %w", err)
	}

	for _, file := range files {
		// formato: file '../tmp/arquivo.ts' ou caminho absoluto/relativo direto se rodado no mesmo contexto
		// O `-safe 0` e paths relativos exigem cuidado. Vamos usar filepath direto.
		absPath, err := filepath.Abs(file)
		if err != nil {
			_ = f.Close()
			return fmt.Errorf("erro ao obter caminho absoluto: %w", err)
		}
		if _, err := f.WriteString(fmt.Sprintf("file '%s'\n", absPath)); err != nil {
			_ = f.Close()
			return fmt.Errorf("erro ao escrever no arquivo list.txt: %w", err)
		}
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("erro ao fechar list.txt: %w", err)
	}

	var cmd *exec.Cmd
	// 37. Adicionar suporte à injeção de uma trilha de áudio no vídeo final
	if audioPath != "" {
		// #nosec G204
		cmd = exec.Command("ffmpeg", "-y", "-f", "concat", "-safe", "0", "-i", listPath, "-i", audioPath, "-c:v", "copy", "-c:a", "aac", "-map", "0:v", "-map", "1:a", "-shortest", finalOutput)
	} else {
		// #nosec G204
		cmd = exec.Command("ffmpeg", "-y", "-f", "concat", "-safe", "0", "-i", listPath, "-c", "copy", finalOutput)
	}

	// Log da execução do concat
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("erro no ffmpeg concat: %w\nOutput: %s", err, string(output))
	}

	return nil
}
