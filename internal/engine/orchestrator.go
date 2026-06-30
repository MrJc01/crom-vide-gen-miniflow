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
	"strings"
	"videogen/internal/models"
	"videogen/internal/utils"
)

type RenderResult struct {
	CardID string
	Path   string
	Err    error
}

// Worker Pool Pattern
func renderWorker(ctx context.Context, id int, jobs <-chan models.Card, results chan<- RenderResult, tmpl models.Template, wg *sync.WaitGroup, renderer Renderer) {
	defer wg.Done()
	for card := range jobs {
		select {
		case <-ctx.Done():
			slog.Warn("Cancelando render worker", "worker", id)
			return
		default:
		}

		outPath := filepath.Join("tmp", fmt.Sprintf("%s.mp4", card.ID))
		slog.Info("Renderizando card", "worker", id, "card", card.ID, "outPath", outPath)

		start := time.Now()
		err := renderer.RenderCard(ctx, card, tmpl.Resolution, tmpl.FPS, outPath)
		elapsed := time.Since(start)

		if err != nil {
			slog.Error("Erro ao renderizar card", "card", card.ID, "erro", err, "duracao_ms", elapsed.Milliseconds())
			results <- RenderResult{CardID: card.ID, Path: "", Err: fmt.Errorf("card %s failed: %v", card.ID, err)}
			continue
		}
		
		slog.Info("Card renderizado com sucesso", "card", card.ID, "duracao_ms", elapsed.Milliseconds())
		results <- RenderResult{CardID: card.ID, Path: outPath, Err: nil}
	}
}

// ProcessVideo Orquestra as chamadas
func ProcessVideo(ctx context.Context, tmpl models.Template, finalOutput string, numWorkers int, renderer Renderer) error {
	var wg sync.WaitGroup

	for i, card := range tmpl.Cards {
		for j, el := range card.Elements {
			if el.Type == "image" && (strings.HasPrefix(el.Content, "http://") || strings.HasPrefix(el.Content, "https://")) {
				slog.Info("Baixando imagem remota do elemento", "url", el.Content)
				localPath, err := utils.DownloadRemoteFile(el.Content, "tmp")
				if err != nil {
					slog.Warn("Falha ao baixar imagem remota (ignorando)", "url", el.Content, "erro", err)
				} else {
					tmpl.Cards[i].Elements[j].Content = localPath
				}
			}
		}
	}

	jobs := make(chan models.Card, len(tmpl.Cards))
	results := make(chan RenderResult, len(tmpl.Cards))

	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go renderWorker(ctx, w, jobs, results, tmpl, &wg, renderer)
	}

	for _, card := range tmpl.Cards {
		jobs <- card
	}
	close(jobs)

	wg.Wait()
	close(results)

	renderedPaths := make(map[string]string)
	var errs []string
	for res := range results {
		if res.Err != nil {
			errs = append(errs, res.Err.Error())
		} else if res.Path != "" {
			renderedPaths[res.CardID] = res.Path
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("falha na renderização: %s", strings.Join(errs, "; "))
	}

	// Rebuild the tmpFiles slice in the exact chronological order of the template cards
	var tmpFiles []string
	for _, card := range tmpl.Cards {
		path, exists := renderedPaths[card.ID]
		if !exists {
			return fmt.Errorf("caminho de renderização para o card %s não encontrado", card.ID)
		}
		tmpFiles = append(tmpFiles, path)
	}

	var audioPath string
	if tmpl.AudioURL != "" {
		if strings.HasPrefix(tmpl.AudioURL, "http://") || strings.HasPrefix(tmpl.AudioURL, "https://") {
			slog.Info("Trilha sonora especificada (remota). Baixando áudio...", "url", tmpl.AudioURL)
			var err error
			audioPath, err = utils.DownloadRemoteFile(tmpl.AudioURL, "tmp")
			if err != nil {
				slog.Warn("Falha ao baixar áudio remoto (ignorando)", "url", tmpl.AudioURL, "erro", err)
				audioPath = ""
			}
		} else {
			slog.Info("Trilha sonora especificada (local). Usando arquivo diretamente...", "path", tmpl.AudioURL)
			if _, err := os.Stat(tmpl.AudioURL); os.IsNotExist(err) {
				slog.Warn("Arquivo de áudio local não encontrado (ignorando)", "path", tmpl.AudioURL)
				audioPath = ""
			} else {
				audioPath = tmpl.AudioURL
			}
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
	// 37. Adicionar suporte à injeção e mixagem de trilha de áudio de fundo
	if audioPath != "" {
		// #nosec G204
		cmd = exec.Command("ffmpeg", "-y", "-f", "concat", "-safe", "0", "-i", listPath, "-i", audioPath,
			"-filter_complex", "[0:a][1:a]amix=inputs=2:duration=first:dropout_transition=2[a]",
			"-map", "0:v", "-map", "[a]", "-c:v", "copy", "-c:a", "aac", finalOutput)
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
