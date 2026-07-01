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
	"path/filepath"
	"syscall"
	"videogen/internal/assets"
	"videogen/internal/engine"
	"videogen/internal/models"
	"videogen/internal/utils"
)

var version = "v0.1.0"

func main() {
	// Extrair recursos embutidos padrões (bootstrapping)
	if err := assets.ExtractAll(); err != nil {
		slog.Warn("Aviso: erro ao extrair recursos embutidos", "erro", err)
	}

	// Prepara PATH local para encontrar o ffmpeg
	if absBin, err := filepath.Abs("bin"); err == nil {
		os.Setenv("PATH", absBin+string(os.PathListSeparator)+os.Getenv("PATH"))
	}

	if err := run(); err != nil {
		slog.Error("Erro fatal na execução", "erro", err)
		os.Exit(1)
	}
}

func run() error {
	// 3. Interface CLI (flag) e 16. Versionamento
	// 43. Parametrizar concorrência via flag CLI
	numWorkers := flag.Int("workers", 0, "Número de workers concorrentes (0 usa runtime.NumCPU ou ENV)")
	jsonPath := flag.String("json", "templates/sample_v1.json", "Caminho do arquivo JSON de template")
	outPath := flag.String("out", "output_final.mp4", "Caminho e nome do arquivo de saída")
	showVersion := flag.Bool("version", false, "Exibe a versão do sistema")
	hwaccel := flag.Bool("hwaccel", false, "Habilita aceleração por hardware (ex: NVENC)")
	printSchema := flag.Bool("schema", false, "Exibe o esquema textual formatado do template e encerra")
	subtitles := flag.Bool("subtitles", true, "Habilita ou desabilita as legendas da narração no vídeo")
	
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
		return nil
	}

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			return fmt.Errorf("falha ao criar arquivo de CPU profile: %w", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			_ = f.Close()
			return fmt.Errorf("falha ao iniciar CPU profile: %w", err)
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
		return fmt.Errorf("falha ao criar diretório temporário: %w", err)
	}
	// 63. Limpeza pós-execução (garantido pela saída da função run())
	defer func() {
		if err := utils.CleanupTempFiles("tmp"); err != nil {
			slog.Warn("Falha ao limpar arquivos temporários", "erro", err)
		}
	}()

	slog.Info("Iniciando videogen CLI", "json", *jsonPath, "out", *outPath)

	// 21 e 22. Checar presença do FFmpeg
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return fmt.Errorf("FFmpeg não encontrado no PATH do sistema. Por favor, instale o FFmpeg para rodar o videogen")
	}

	file, err := os.ReadFile(*jsonPath)
	if err != nil {
		return fmt.Errorf("erro ao ler JSON: %w", err)
	}

	var template models.Template
	if err := json.Unmarshal(file, &template); err != nil {
		return fmt.Errorf("erro ao fazer parse do JSON: %w", err)
	}

	// Expande os templates de cena pré-configurados
	engine.ExpandTemplates(&template)

	absJsonPath, err := filepath.Abs(*jsonPath)
	if err != nil {
		return fmt.Errorf("erro ao obter caminho absoluto do JSON: %w", err)
	}
	workspaceDir := filepath.Dir(absJsonPath)

	engine.ResolveRelativePaths(&template, workspaceDir)

	if err := engine.ResolveNarrationAndDurations(ctx, &template); err != nil {
		return fmt.Errorf("erro ao processar narração e durações: %w", err)
	}

	if err := template.Validate(); err != nil {
		return fmt.Errorf("JSON inválido: %w", err)
	}

	slog.Info("JSON parseado e validado com sucesso", "template_id", template.TemplateID)

	if *printSchema {
		fmt.Println(template.GenerateSchemaPrint())
		return nil
	}

	// 4. Estruturar a camada de configuração (leitura de variáveis de ambiente)
	// 44. Adaptar workers aos cores lógicos ou à env
	workersCount := 0
	if *numWorkers > 0 {
		workersCount = *numWorkers
	} else if envWorkers := os.Getenv("VIDEOGEN_WORKERS"); envWorkers != "" {
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

	showSubtitles := true
	if template.Subtitles != nil {
		showSubtitles = *template.Subtitles
	}
	isSubtitlesSet := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "subtitles" {
			isSubtitlesSet = true
		}
	})
	if isSubtitlesSet {
		showSubtitles = *subtitles
	}

	// 14. Estruturar camada de injeção de dependências
	renderer := engine.NewFFmpegRenderer(*hwaccel, template.JPEGQuality, showSubtitles)

	err = engine.ProcessVideo(ctx, template, *outPath, workersCount, renderer)
	if err != nil {
		return fmt.Errorf("falha crítica ao gerar vídeo: %w", err)
	}

	slog.Info("Setup e execução concluídos com sucesso!", "video_path", *outPath)

	// Simula a escuta do graceful shutdown
	select {
	case <-ctx.Done():
		slog.Info("Desligamento gracioso acionado. Limpando recursos...")
	default:
	}

	return nil
}
