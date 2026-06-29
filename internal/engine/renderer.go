package engine

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"image"
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
func (r *FFmpegRenderer) RenderCard(ctx context.Context, card models.Card, res models.Size, fps int, outPath string) (err error) {
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
	vfFilter := fmt.Sprintf("fade=t=in:start_frame=0:nb_frames=%d,fade=t=out:start_frame=%d:nb_frames=%d", fadeFrames, totalFrames-fadeFrames, fadeFrames)

	// 29. Comando FFmpeg esperando imagens no Stdin (Pipe) via rawvideo
	// #nosec G204
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-y",
		"-f", "rawvideo",
		"-pix_fmt", "rgba",
		"-s", fmt.Sprintf("%dx%d", res.Width, res.Height),
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
	dc := gg.NewContext(res.Width, res.Height)

	// 46. Usar buffers redimensionáveis para codificação (sync.Pool)
	bw := bufWriterPool.Get().(*bufio.Writer)
	bw.Reset(stdin)

	// Usamos named return value (err) para capturar o erro do Wait() no defer de forma limpa e evitar processos zumbis
	defer func() {
		_ = stdin.Close()
		bufWriterPool.Put(bw)
		if cmd.Process != nil {
			waitErr := cmd.Wait()
			if waitErr != nil {
				if err != nil {
					err = fmt.Errorf("%w (log do ffmpeg: %s)", err, stderr.String())
				} else {
					// 40. Mapear e tratar erros do FFmpeg com logs legíveis
					err = fmt.Errorf("ffmpeg falhou: %w, log do ffmpeg: %s", waitErr, stderr.String())
				}
			}
		}
	}()

	// Loop desenhando os frames na RAM
	for f := 0; f < totalFrames; f++ {
		// 45. Otimizar alocação de memória no loop de desenho de frames (GC)
		DrawCardState(dc, card, res)

		// Conversão direta para image.RGBA para extração de pixels sem compressão no Go (Zero-copy pixel stream)
		rgbaImg, ok := dc.Image().(*image.RGBA)
		if !ok {
			err = fmt.Errorf("imagem gerada não é do tipo *image.RGBA")
			return err
		}

		// Envia os pixels crus diretamente pro FFmpeg via pipe
		if _, writeErr := bw.Write(rgbaImg.Pix); writeErr != nil {
			err = fmt.Errorf("erro ao escrever pixels no pipe: %w", writeErr)
			return err
		}
		_ = bw.Flush() // Garante envio imediato ao FFmpeg subprocess
	}

	_ = bw.Flush()
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

			// Escalamento responsivo do tamanho do texto baseando-se em largura de 1080p de referência
			scale := float64(res.Width) / 1080.0
			scaledFontSize := el.FontSize * scale
			if scaledFontSize < 10 {
				scaledFontSize = 10 // Limite mínimo de legibilidade
			}

			// 25. Carregar fonte customizada (stub com fallback)
			if err := dc.LoadFontFace("assets/fonts/Roboto.ttf", scaledFontSize); err != nil {
				// Fallback silencioso se não houver fonte (graceful degradation)
			}

			// 26. Implementar quebra de linha (Word Wrap) responsiva (margem lateral de 15%)
			maxWidth := float64(res.Width) * 0.85
			dc.DrawStringWrapped(el.Content, el.X, el.Y, 0.5, 0.5, maxWidth, 1.5, gg.AlignCenter)
		} else if el.Type == "image" {
			// 27. Suporte a renderização de imagens estáticas sobrepostas
			img, err := gg.LoadImage(el.Content)
			if err != nil {
				// Fallback gracioso se a imagem não carregar
				continue
			}
			dc.DrawImageAnchored(img, int(el.X), int(el.Y), 0.5, 0.5)
		}
	}
}
