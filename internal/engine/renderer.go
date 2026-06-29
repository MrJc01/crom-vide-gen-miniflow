package engine

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"videogen/internal/models"

	"github.com/fogleman/gg"
	_ "image/jpeg"
	_ "image/png"
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

	// Crias diretórios temporários para extração de frames de elementos de vídeo se houver
	for i, el := range card.Elements {
		if el.Type == "video" {
			framesDir := fmt.Sprintf("tmp/frames_%s_%d", card.ID, i)
			_ = os.MkdirAll(framesDir, 0755)

			w := int(el.Width)
			if w <= 0 {
				w = 400
			}
			h := int(el.Height)
			if h <= 0 {
				h = 400
			}

			// Check if file exists or is valid before running ffmpeg
			if el.Content == "" {
				slog.Warn("Vídeo sem arquivo (vazio), ignorando extração", "card", card.ID)
				continue
			}
			if !strings.HasPrefix(el.Content, "http://") && !strings.HasPrefix(el.Content, "https://") {
				if _, err := os.Stat(el.Content); os.IsNotExist(err) {
					slog.Warn("Vídeo local não encontrado, ignorando extração", "file", el.Content)
					continue
				}
			}

			// Extrai frames do vídeo no formato frame_%04d.jpg na resolução e fps apropriados
			// #nosec G204
			extractCmd := exec.CommandContext(ctx, "ffmpeg", "-y",
				"-t", fmt.Sprintf("%.3f", float64(card.DurationMs)/1000.0),
				"-i", el.Content,
				"-vf", fmt.Sprintf("fps=%d,scale=%d:%d", fps, w, h),
				"-q:v", "2",
				filepath.Join(framesDir, "frame_%04d.jpg"),
			)
			out, err := extractCmd.CombinedOutput()
			if err != nil {
				slog.Warn("Falha ao extrair frames (ignorando vídeo)", "file", el.Content, "erro", err, "output", string(out))
				continue
			}

			// Tenta extrair áudio se houver trilha sonora no vídeo
			audioFile := fmt.Sprintf("tmp/audio_%s_%d.aac", card.ID, i)
			// #nosec G204
			audioCmd := exec.CommandContext(ctx, "ffmpeg", "-y",
				"-t", fmt.Sprintf("%.3f", float64(card.DurationMs)/1000.0),
				"-i", el.Content,
				"-vn",
				"-c:a", "aac",
				audioFile,
			)
			_ = audioCmd.Run() // Ignora erro se o vídeo não tiver áudio
		}
	}

	bitrate := "2M"
	if res.Width >= 1920 || res.Height >= 1920 {
		bitrate = "8M"
	} else if res.Width >= 1280 || res.Height >= 1280 {
		bitrate = "4M"
	}

	vcodec := "libx264"
	if r.UseHWAccel {
		vcodec = "h264_nvenc"
	}

	fadeFrames := fps / 2
	if fadeFrames < 1 {
		fadeFrames = 1
	}
	vfFilter := fmt.Sprintf("fade=t=in:start_frame=0:nb_frames=%d,fade=t=out:start_frame=%d:nb_frames=%d", fadeFrames, totalFrames-fadeFrames, fadeFrames)

	// Comando FFmpeg esperando imagens no Stdin (Pipe) via rawvideo e adicionando anullsrc para canal de áudio silencioso contínuo
	// #nosec G204
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-y",
		"-f", "rawvideo",
		"-pix_fmt", "rgba",
		"-s", fmt.Sprintf("%dx%d", res.Width, res.Height),
		"-r", fmt.Sprintf("%d", fps),
		"-i", "-",
		"-f", "lavfi",
		"-i", "anullsrc=channel_layout=stereo:sample_rate=44100",
		"-c:v", vcodec,
		"-b:v", bitrate,
		"-vf", vfFilter,
		"-c:a", "aac",
		"-shortest",
		"-preset", "veryfast",
		"-pix_fmt", "yuv420p",
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

	dc := gg.NewContext(res.Width, res.Height)

	bw := bufWriterPool.Get().(*bufio.Writer)
	bw.Reset(stdin)

	// O retorno nomeado (err) garante que erros do Wait() atualizem o retorno da chamada
	defer func() {
		// Cleanup the buffer if needed
		bufWriterPool.Put(bw)

		// Remove os diretórios temporários dos frames após a execução
		for i, el := range card.Elements {
			if el.Type == "video" {
				framesDir := fmt.Sprintf("tmp/frames_%s_%d", card.ID, i)
				_ = os.RemoveAll(framesDir)
			}
		}
	}()

	for f := 0; f < totalFrames; f++ {
		DrawCardState(dc, card, res, f)

		rgbaImg, ok := dc.Image().(*image.RGBA)
		if !ok {
			err = fmt.Errorf("imagem gerada não é do tipo *image.RGBA")
			return err
		}

		if _, writeErr := bw.Write(rgbaImg.Pix); writeErr != nil {
			err = fmt.Errorf("erro ao escrever pixels no pipe: %w", writeErr)
			return err
		}
		_ = bw.Flush()
	}

	_ = bw.Flush()

	// Finaliza a escrita de pixels no ffmpeg, forçando-o a processar e gerar o arquivo de saída (.ts)
	_ = stdin.Close()
	if cmd.Process != nil {
		waitErr := cmd.Wait()
		if waitErr != nil {
			if err != nil {
				err = fmt.Errorf("%w (log do ffmpeg: %s)", err, stderr.String())
			} else {
				err = fmt.Errorf("ffmpeg falhou: %w, log do ffmpeg: %s", waitErr, stderr.String())
			}
			return err
		}
	}

	// Se houver um áudio extraído, faz o mux dele com o arquivo .ts silencioso recém-gerado!
	var audioToMerge string
	for i, el := range card.Elements {
		if el.Type == "video" {
			audioFile := fmt.Sprintf("tmp/audio_%s_%d.aac", card.ID, i)
			if info, err := os.Stat(audioFile); err == nil && info.Size() > 100 {
				audioToMerge = audioFile
				break // Pega o primeiro áudio com som
			}
		}
	}

	if audioToMerge != "" {
		tempMp4 := outPath + ".temp.mp4"
		// #nosec G204
		mergeCmd := exec.CommandContext(ctx, "ffmpeg", "-y",
			"-i", outPath,
			"-i", audioToMerge,
			"-c:v", "copy",
			"-c:a", "aac",
			"-map", "0:v:0",
			"-map", "1:a:0",
			"-shortest",
			tempMp4,
		)
		if err := mergeCmd.Run(); err == nil {
			_ = os.Rename(tempMp4, outPath)
		} else {
			_ = os.Remove(tempMp4)
		}
	}

	// Remove os arquivos temporários de áudio
	for i, el := range card.Elements {
		if el.Type == "video" {
			_ = os.Remove(fmt.Sprintf("tmp/audio_%s_%d.aac", card.ID, i))
		}
	}

	return nil
}

func parseHexColor(s string) color.Color {
	s = strings.TrimPrefix(s, "#")
	c := color.RGBA{A: 255}
	if len(s) == 6 {
		if r, err := strconv.ParseUint(s[0:2], 16, 8); err == nil { c.R = uint8(r) }
		if g, err := strconv.ParseUint(s[2:4], 16, 8); err == nil { c.G = uint8(g) }
		if b, err := strconv.ParseUint(s[4:6], 16, 8); err == nil { c.B = uint8(b) }
	} else if len(s) == 8 {
		if r, err := strconv.ParseUint(s[0:2], 16, 8); err == nil { c.R = uint8(r) }
		if g, err := strconv.ParseUint(s[2:4], 16, 8); err == nil { c.G = uint8(g) }
		if b, err := strconv.ParseUint(s[4:6], 16, 8); err == nil { c.B = uint8(b) }
		if a, err := strconv.ParseUint(s[6:8], 16, 8); err == nil { c.A = uint8(a) }
	}
	return c
}

// DrawCardState desenha um quadro específico de um Card
func DrawCardState(dc *gg.Context, card models.Card, res models.Size, frameIndex int) {
	// Fundo
	dc.SetHexColor(card.BackgroundColor)
	dc.Clear()

	for elIdx, el := range card.Elements {
		dc.Push()
		if el.Rotation != 0 {
			dc.RotateAbout(gg.Radians(el.Rotation), el.X, el.Y)
		}

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
			align := gg.AlignCenter
			ax, ay := 0.5, 0.5
			if el.TextAlign == "left" {
				align = gg.AlignLeft
				ax = 0.0
			} else if el.TextAlign == "right" {
				align = gg.AlignRight
				ax = 1.0
			}
			if el.ShadowColor != "" {
				// Drop shadow (hard shadow based on offset)
				dc.SetHexColor(el.ShadowColor)
				dc.DrawStringWrapped(el.Content, el.X+el.ShadowOffsetX, el.Y+el.ShadowOffsetY, ax, ay, maxWidth, 1.5, align)
				// Revert color for main text
				dc.SetHexColor(el.Color)
			}
			dc.DrawStringWrapped(el.Content, el.X, el.Y, ax, ay, maxWidth, 1.5, align)
		} else if el.Type == "rect" {
			// Suporte a renderização de retângulos/banners com gradientes 3D
			if strings.HasPrefix(el.Color, "gradient:") {
				colors := strings.Split(strings.TrimPrefix(el.Color, "gradient:"), ",")
				if len(colors) == 2 {
					grad := gg.NewLinearGradient(el.X, el.Y-el.Height/2, el.X, el.Y+el.Height/2)
					// Helper básico de hex local se necessário, mas vamos assumir que ParseHexColor não falhe
					c1 := parseHexColor(colors[0])
					c2 := parseHexColor(colors[1])
					grad.AddColorStop(0, c1)
					grad.AddColorStop(1, c2)
					dc.SetFillStyle(grad)
				}
			} else {
				dc.SetHexColor(el.Color)
			}
			dc.DrawRectangle(el.X-el.Width/2, el.Y-el.Height/2, el.Width, el.Height)
			dc.Fill()
		} else if el.Type == "polygon" {
			if len(el.Points) > 2 {
				dc.SetHexColor(el.Color)
				dc.MoveTo(el.Points[0][0], el.Points[0][1])
				for i := 1; i < len(el.Points); i++ {
					dc.LineTo(el.Points[i][0], el.Points[i][1])
				}
				dc.ClosePath()
				dc.Fill()
			}
		} else if el.Type == "circle" {
			// Suporte a círculos (ex: dot de notificação/gravação)
			dc.SetHexColor(el.Color)
			dc.DrawCircle(el.X, el.Y, el.Width)
			dc.Fill()
		} else if el.Type == "frame" {
			// Suporte a molduras vazadas
			dc.SetHexColor(el.Color)
			strokeWidth := el.StrokeWidth
			if strokeWidth <= 0 {
				strokeWidth = 5.0
			}
			dc.SetLineWidth(strokeWidth)
			dc.DrawRectangle(el.X-el.Width/2, el.Y-el.Height/2, el.Width, el.Height)
			dc.Stroke()
		} else if el.Type == "image" {
			// 27. Suporte a renderização de imagens estáticas sobrepostas
			img, err := gg.LoadImage(el.Content)
			if err != nil {
				dc.Pop()
				continue
			}
			dc.DrawImageAnchored(img, int(el.X), int(el.Y), 0.5, 0.5)
		} else if el.Type == "video" {
			// 27. Suporte a renderização de elementos de vídeo frame a frame
			framePath := fmt.Sprintf("tmp/frames_%s_%d/frame_%04d.jpg", card.ID, elIdx, frameIndex+1)
			if _, err := os.Stat(framePath); err != nil {
				// Fallback loop/hold: se o vídeo acabou, tenta repetir o primeiro frame
				framePath = fmt.Sprintf("tmp/frames_%s_%d/frame_0001.jpg", card.ID, elIdx)
			}

			img, err := gg.LoadImage(framePath)
			if err != nil {
				// Fallback visual (Placeholder) para Web Preview
				dc.SetHexColor("#333333")
				dc.DrawRectangle(el.X-el.Width/2, el.Y-el.Height/2, el.Width, el.Height)
				dc.Fill()
				
				dc.SetHexColor("#FFFFFF")
				dc.DrawStringAnchored("VIDEO PLACEHOLDER", el.X, el.Y, 0.5, 0.5)
				dc.Pop()
				continue
			}
			dc.DrawImageAnchored(img, int(el.X), int(el.Y), 0.5, 0.5)
		}
		
		dc.Pop()
	}
}
