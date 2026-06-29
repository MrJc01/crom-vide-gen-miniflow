package engine

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"image"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
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

			// Extrai frames do vídeo no formato frame_%04d.png na resolução e fps apropriados
			// #nosec G204
			extractCmd := exec.CommandContext(ctx, "ffmpeg", "-y",
				"-i", el.Content,
				"-vf", fmt.Sprintf("fps=%d,scale=%d:%d", fps, w, h),
				filepath.Join(framesDir, "frame_%04d.png"),
			)
			if out, err := extractCmd.CombinedOutput(); err != nil {
				slog.Warn("Falha ao extrair frames do vídeo", "video", el.Content, "erro", err, "output", string(out))
			}

			// Tenta extrair áudio se houver trilha sonora no vídeo
			audioFile := fmt.Sprintf("tmp/audio_%s_%d.aac", card.ID, i)
			// #nosec G204
			audioCmd := exec.CommandContext(ctx, "ffmpeg", "-y",
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
		_ = stdin.Close()
		bufWriterPool.Put(bw)

		if cmd.Process != nil {
			waitErr := cmd.Wait()
			if waitErr != nil {
				if err != nil {
					err = fmt.Errorf("%w (log do ffmpeg: %s)", err, stderr.String())
				} else {
					err = fmt.Errorf("ffmpeg falhou: %w, log do ffmpeg: %s", waitErr, stderr.String())
				}
			}
		}

		// Remove os diretórios temporários dos frames após a execução do ffmpeg Wait
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
		tempTs := outPath + ".temp.ts"
		// #nosec G204
		mergeCmd := exec.CommandContext(ctx, "ffmpeg", "-y",
			"-i", outPath,
			"-i", audioToMerge,
			"-c:v", "copy",
			"-c:a", "aac",
			"-map", "0:v:0",
			"-map", "1:a:0",
			"-shortest",
			tempTs,
		)
		if err := mergeCmd.Run(); err == nil {
			_ = os.Rename(tempTs, outPath)
		} else {
			_ = os.Remove(tempTs)
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

// DrawCardState desenha o estado atual do card no contexto do GG
func DrawCardState(dc *gg.Context, card models.Card, res models.Size, frameIndex int) {
	// Fundo
	dc.SetHexColor(card.BackgroundColor)
	dc.Clear()

	for elIdx, el := range card.Elements {
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
				continue
			}
			dc.DrawImageAnchored(img, int(el.X), int(el.Y), 0.5, 0.5)
		} else if el.Type == "video" {
			// 27. Suporte a renderização de elementos de vídeo frame a frame
			framePath := fmt.Sprintf("tmp/frames_%s_%d/frame_%04d.png", card.ID, elIdx, frameIndex+1)
			if _, err := os.Stat(framePath); err != nil {
				// Fallback loop/hold: se o vídeo acabou, tenta repetir o primeiro frame
				framePath = fmt.Sprintf("tmp/frames_%s_%d/frame_0001.png", card.ID, elIdx)
			}

			img, err := gg.LoadImage(framePath)
			if err != nil {
				continue
			}
			dc.DrawImageAnchored(img, int(el.X), int(el.Y), 0.5, 0.5)
		}
	}
}
