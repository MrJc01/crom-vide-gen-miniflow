package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"
	"videogen/internal/engine"
	"videogen/internal/models"
	"videogen/internal/utils"
)

var version = "v0.1.0"

func main() {
	// 3. Interface CLI (flag) e 16. Versionamento
	// 43. Parametrizar concorrência via flag CLI
	numWorkers := flag.Int("workers", 0, "Número de workers concorrentes (0 usa runtime.NumCPU ou ENV)")
	jsonPath := flag.String("json", "templates/sample_v1.json", "Caminho do arquivo JSON de template")
	outPath := flag.String("out", "output_final.mp4", "Caminho e nome do arquivo de saída")
	showVersion := flag.Bool("version", false, "Exibe a versão do sistema")
	hwaccel := flag.Bool("hwaccel", false, "Habilita aceleração por hardware (ex: NVENC)")
	
	// 49. Profiling de CPU e Memória via pprof
	cpuprofile := flag.String("cpuprofile", "", "Grava CPU profile no arquivo especificado")
	memprofile := flag.String("memprofile", "", "Grava Memory profile no arquivo especificado")
	
	flag.Parse()

	// 4. Estruturar a camada de configuração (leitura de variáveis de ambiente com prioridade inferior às flags)
	if *jsonPath == "templates/sample_v1.json" {
		if envJson := os.Getenv("VIDEOGEN_JSON"); envJson != "" {
			*jsonPath = envJson
		}
	}
	if *outPath == "output_final.mp4" {
		if envOut := os.Getenv("VIDEOGEN_OUT"); envOut != "" {
			*outPath = envOut
		}
	}

	if *showVersion {
		_, _ = os.Stdout.WriteString("videogen version " + version + "\n")
		return
	}

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			slog.Error("Falha ao criar arquivo de CPU profile", "erro", err)
			os.Exit(1)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			slog.Error("Falha ao iniciar CPU profile", "erro", err)
			os.Exit(1)
		}
		defer pprof.StopCPUProfile()
	}

	// Agendar dump do Heap profile para o final
	if *memprofile != "" {
		defer func() {
			f, err := os.Create(*memprofile)
			if err != nil {
				slog.Error("Falha ao criar arquivo de Memory profile", "erro", err)
				return
			}
			if err := pprof.WriteHeapProfile(f); err != nil {
				slog.Error("Falha ao gravar Memory profile", "erro", err)
			}
			_ = f.Close()
		}()
	}

	// 10. Logging estruturado
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// 15. Graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// 11. Criar gerenciador de arquivos temporários e diretório de saída
	if err := utils.EnsureDirectories("tmp"); err != nil {
		slog.Error("Falha ao criar diretório temporário", "erro", err)
		os.Exit(1)
	}
	// 63. Limpeza pós-execução
	defer func() {
		if err := utils.CleanupTempFiles("tmp"); err != nil {
			slog.Warn("Falha ao limpar arquivos temporários", "erro", err)
		}
	}()

	slog.Info("Iniciando videogen CLI", "json", *jsonPath, "out", *outPath)

	// 21 e 22. Checar presença do FFmpeg
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		slog.Error("FFmpeg não encontrado no PATH do sistema. Por favor, instale o FFmpeg para rodar o videogen.")
		os.Exit(1)
	}

	file, err := os.ReadFile(*jsonPath)
	if err != nil {
		slog.Error("Erro ao ler JSON", "erro", err)
		os.Exit(1)
	}

	var template models.Template
	if err := json.Unmarshal(file, &template); err != nil {
		slog.Error("Erro ao fazer parse do JSON", "erro", err)
		os.Exit(1)
	}

	if err := template.Validate(); err != nil {
		slog.Error("JSON invalido", "erro", err)
		os.Exit(1)
	}

	slog.Info("JSON parseado e validado com sucesso", "template_id", template.TemplateID)

	// 4. Estruturar a camada de configuração (leitura de variáveis de ambiente)
	// 44. Adaptar workers aos cores lógicos ou à env
	workersCount := 0
	if *numWorkers > 0 {
		workersCount = *numWorkers
	} else if envWorkers := os.Getenv("VIDEOGEN_WORKERS"); envWorkers != "" {
		// Conversão rápida de string para int
		var parsed int
		_, _ = fmt.Sscanf(envWorkers, "%d", &parsed)
		if parsed > 0 {
			workersCount = parsed
		}
	}
	if workersCount <= 0 {
		workersCount = runtime.NumCPU()
	}

	slog.Info("Configuração do pool de concorrência", "workers", workersCount)

	// 14. Estruturar camada de injeção de dependências
	renderer := engine.NewFFmpegRenderer(*hwaccel)

	err = engine.ProcessVideo(ctx, template, *outPath, workersCount, renderer)
	if err != nil {
		slog.Error("Falha crítica ao gerar vídeo", "erro", err)
		os.Exit(1)
	}

	slog.Info("Setup e execução concluídos com sucesso!", "video_path", *outPath)

	// Simula a escuta do graceful shutdown
	select {
	case <-ctx.Done():
		slog.Info("Desligamento gracioso acionado. Limpando recursos...")
		// Limpeza iria aqui
	default:
		// Continua o fluxo
	}
}
