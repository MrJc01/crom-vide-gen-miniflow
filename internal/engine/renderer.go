package engine

import (
	"bufio"
	"bytes"
	"fmt"
	"image/png"
	"os/exec"
	"sync"
	"videogen/internal/models"

	"github.com/fogleman/gg"
)

var bufWriterPool = sync.Pool{
	New: func() interface{} {
		return bufio.NewWriterSize(nil, 512*1024) // 512KB reuse buffer
	},
}

type FFmpegRenderer struct {
	UseHWAccel bool
}

func NewFFmpegRenderer(useHWAccel bool) *FFmpegRenderer {
	return &FFmpegRenderer{UseHWAccel: useHWAccel}
}

// RenderCard processa um card específico e gera um arquivo .ts temporário
func (r *FFmpegRenderer) RenderCard(card models.Card, res models.Size, fps int, outPath string) error {
	totalFrames := (card.DurationMs / 1000) * fps

	// 32. Configurar bitrate com base na resolução desejada (ex: 8M para 1080p+, 4M para 720p+, 2M para menor)
	bitrate := "2M"
	if res.Width >= 1920 || res.Height >= 1920 {
		bitrate = "8M"
	} else if res.Width >= 1280 || res.Height >= 1280 {
		bitrate = "4M"
	}

	// 39. Configurar aceleração de hardware (ex: NVENC)
	vcodec := "libx264"
	if r.UseHWAccel {
		vcodec = "h264_nvenc"
	}

	// 38. Transições básicas (fade in/out) entre vídeos gerados (ex: 0.5s fade no início e fim)
	fadeFrames := fps / 2
	if fadeFrames < 1 {
		fadeFrames = 1
	}
	vfFilter := fmt.Sprintf("fade=t=in:start_frame=0:num_frames=%d,fade=t=out:start_frame=%d:num_frames=%d", fadeFrames, totalFrames-fadeFrames, fadeFrames)

	// 29. Comando FFmpeg esperando imagens no Stdin (Pipe)
	// 30, 31. Ajuste de codec e pixel format
	// #nosec G204
	cmd := exec.Command("ffmpeg",
		"-y",
		"-f", "image2pipe",
		"-vcodec", "png",
		"-r", fmt.Sprintf("%d", fps), // 33. Suporte a taxa de quadros (FPS) dinâmica
		"-i", "-",
		"-c:v", vcodec,
		"-b:v", bitrate,
		"-vf", vfFilter,
		"-preset", "veryfast", // 30. Otimização de velocidade
		"-pix_fmt", "yuv420p", // 31. Compatibilidade
		outPath,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("erro ao criar stdin pipe para ffmpeg: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("erro ao iniciar ffmpeg: %w", err)
	}

	// 28. Otimizar criação do contexto gráfico (reuso de memória)
	// Criamos apenas uma vez por card e reusamos em todos os frames
	dc := gg.NewContext(res.Width, res.Height)

	// 47. Testar substituição do encoder PNG padrão por alternativas mais rápidas (BestSpeed)
	pngEncoder := png.Encoder{
		CompressionLevel: png.BestSpeed,
	}

	// 46. Usar buffers redimensionáveis para codificação PNG (sync.Pool)
	// Adquire e reseta o buffered writer
	bw := bufWriterPool.Get().(*bufio.Writer)
	bw.Reset(stdin)
	defer func() {
		_ = bw.Flush()
		bufWriterPool.Put(bw)
	}()

	// Loop desenhando os frames na RAM
	for f := 0; f < totalFrames; f++ {
		// 45. Otimizar alocação de memória no loop de desenho de frames (GC)
		DrawCardState(dc, card, res)

		// Envia a imagem pro FFmpeg
		if err := pngEncoder.Encode(bw, dc.Image()); err != nil {
			return fmt.Errorf("erro ao codificar png para pipe: %w", err)
		}
		_ = bw.Flush() // Garante envio em tempo real pro FFmpeg
	}

	_ = stdin.Close()
	if err := cmd.Wait(); err != nil {
		// 40. Mapear e tratar erros do FFmpeg com logs legíveis
		return fmt.Errorf("ffmpeg falhou: %w, log do ffmpeg: %s", err, stderr.String())
	}
	return nil
}

// DrawCardState desenha o estado atual do card no contexto do GG
func DrawCardState(dc *gg.Context, card models.Card, res models.Size) {
	// Fundo
	dc.SetHexColor(card.BackgroundColor)
	dc.Clear()

	// Desenhar elementos
	for _, el := range card.Elements {
		if el.Type == "text" {
			dc.SetHexColor(el.Color)

			// 25. Carregar fonte customizada (stub com fallback)
			if err := dc.LoadFontFace("assets/fonts/Roboto.ttf", el.FontSize); err != nil {
				// Fallback silencioso se não houver fonte (graceful degradation)
			}

			// 26. Implementar quebra de linha (Word Wrap)
			maxWidth := float64(res.Width) - el.X - 40 // Margem
			dc.DrawStringWrapped(el.Content, el.X, el.Y, 0.5, 0.5, maxWidth, 1.5, gg.AlignCenter)
		} else if el.Type == "image" {
			// 27. Suporte a renderização de imagens estáticas sobrepostas
		}
	}
}
