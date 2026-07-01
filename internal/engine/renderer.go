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
	"github.com/golang/freetype/truetype"
	_ "image/jpeg"
	_ "image/png"
	"golang.org/x/image/font"
)

var bufWriterPool = sync.Pool{
	New: func() interface{} {
		return bufio.NewWriterSize(nil, 512*1024) // 512KB reuse buffer
	},
}

type FFmpegRenderer struct {
	UseHWAccel    bool
	JPEGQuality   int
	ShowSubtitles bool
}

var (
	nvencSupported bool
	nvencChecked   sync.Once
)

func checkNVENC() bool {
	nvencChecked.Do(func() {
		// Executa um comando rápido e curto de teste para saber se o driver nvenc funciona na máquina local
		cmd := exec.Command("ffmpeg", "-f", "lavfi", "-i", "nullsrc=s=640x480", "-c:v", "h264_nvenc", "-t", "0.01", "-f", "null", "-")
		err := cmd.Run()
		nvencSupported = (err == nil)
		if !nvencSupported {
			slog.Warn("Aceleração por hardware (h264_nvenc) indisponível neste ambiente. O renderizador usará CPU (libx264).")
		} else {
			slog.Info("Aceleração por hardware (h264_nvenc) suportada e pronta para uso!")
		}
	})
	return nvencSupported
}

func NewFFmpegRenderer(useHWAccel bool, jpegQuality int, showSubtitles bool) *FFmpegRenderer {
	if jpegQuality < 1 || jpegQuality > 31 {
		jpegQuality = 2
	}
	return &FFmpegRenderer{
		UseHWAccel:    useHWAccel,
		JPEGQuality:   jpegQuality,
		ShowSubtitles: showSubtitles,
	}
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
				"-q:v", strconv.Itoa(r.JPEGQuality),
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
	if r.UseHWAccel && checkNVENC() {
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
		"-ar", "44100",
		"-ac", "2",
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

	// Pre-load Roboto font to avoid reading/parsing Roboto.ttf on every frame
	var robotoFont *truetype.Font
	if fontBytes, fontErr := os.ReadFile("assets/fonts/Roboto.ttf"); fontErr == nil {
		if f, parseErr := truetype.Parse(fontBytes); parseErr == nil {
			robotoFont = f
		}
	}

	// Pre-load and cache static images to avoid decoding the same image on every frame
	imageCache := make(map[string]image.Image)
	for _, el := range card.Elements {
		if el.Type == "image" && el.Content != "" {
			if _, exists := imageCache[el.Content]; !exists {
				if img, imgErr := gg.LoadImage(el.Content); imgErr == nil {
					imageCache[el.Content] = img
				}
			}
		}
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
		DrawCardState(dc, card, res, f, imageCache, robotoFont, r.ShowSubtitles)

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

	// Se houver um áudio extraído ou TTS, faz o mux dele com o arquivo .ts silencioso recém-gerado!
	var videoAudio string
	for i, el := range card.Elements {
		if el.Type == "video" {
			audioFile := fmt.Sprintf("tmp/audio_%s_%d.aac", card.ID, i)
			if info, err := os.Stat(audioFile); err == nil && info.Size() > 100 {
				videoAudio = audioFile
				break // Pega o primeiro áudio com som
			}
		}
	}

	ttsFile := filepath.Join("tmp", fmt.Sprintf("tts_%s.mp3", card.ID))
	hasTTS := false
	if info, err := os.Stat(ttsFile); err == nil && info.Size() > 100 {
		hasTTS = true
	}

	if hasTTS || videoAudio != "" {
		tempMp4 := outPath + ".temp.mp4"
		var mergeCmd *exec.Cmd
		durSec := fmt.Sprintf("%.3f", float64(card.DurationMs)/1000.0)
		
		if hasTTS && videoAudio != "" {
			// #nosec G204
			mergeCmd = exec.CommandContext(ctx, "ffmpeg", "-y",
				"-i", outPath,
				"-i", ttsFile,
				"-i", videoAudio,
				"-filter_complex", "[1:a]adelay=delays=1000:all=1[delayed_tts];[0:a][delayed_tts][2:a]amix=inputs=3:duration=first:normalize=0[a]",
				"-c:v", "copy",
				"-c:a", "aac",
				"-ar", "44100",
				"-ac", "2",
				"-map", "0:v:0",
				"-map", "[a]",
				"-t", durSec,
				tempMp4,
			)
		} else if hasTTS {
			// #nosec G204
			mergeCmd = exec.CommandContext(ctx, "ffmpeg", "-y",
				"-i", outPath,
				"-i", ttsFile,
				"-filter_complex", "[1:a]adelay=delays=1000:all=1[delayed_tts];[0:a][delayed_tts]amix=inputs=2:duration=first:normalize=0[a]",
				"-c:v", "copy",
				"-c:a", "aac",
				"-ar", "44100",
				"-ac", "2",
				"-map", "0:v:0",
				"-map", "[a]",
				"-t", durSec,
				tempMp4,
			)
		} else {
			// #nosec G204
			mergeCmd = exec.CommandContext(ctx, "ffmpeg", "-y",
				"-i", outPath,
				"-i", videoAudio,
				"-filter_complex", "[0:a][1:a]amix=inputs=2:duration=first:normalize=0[a]",
				"-c:v", "copy",
				"-c:a", "aac",
				"-ar", "44100",
				"-ac", "2",
				"-map", "0:v:0",
				"-map", "[a]",
				"-t", durSec,
				tempMp4,
			)
		}

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

func DrawCardState(dc *gg.Context, card models.Card, res models.Size, frameIndex int, imageCache map[string]image.Image, robotoFont *truetype.Font, showSubtitles bool) {
	// Fundo
	dc.SetHexColor(card.BackgroundColor)
	dc.Clear()

	centerX := float64(res.Width) / 2.0
	centerY := float64(res.Height) / 2.0

	isPortrait := res.Height > res.Width
	scaleX := float64(res.Width) / 1920.0
	scaleY := float64(res.Height) / 1080.0

	refWidth := 1920.0
	if isPortrait {
		refWidth = 1080.0
	}

	for elIdx, el := range card.Elements {
		dc.Push()
		dc.Translate(centerX, centerY)

		// Copia o elemento para aplicar ajustes responsivos sem modificar a estrutura original
		resEl := el

		if card.TemplateName != "" {
			if isPortrait {
				// Ajustes para modo retrato (vertical) nos templates
				switch card.TemplateName {
				case "intro":
					if resEl.Type == "rect" && resEl.Width == 1920 && resEl.Height == 1080 {
						resEl.Width = float64(res.Width)
						resEl.Height = float64(res.Height)
					} else if resEl.Type == "image" { // Logo / Foreground image overlay
						resEl.X = 0
						resEl.Y = -float64(res.Height) * 0.26
						resEl.Width = float64(res.Width) * 0.35
						resEl.Height = float64(res.Width) * 0.35
					} else if resEl.Type == "text" {
						resEl.X = 0
						resEl.TextAlign = "center"
						if resEl.FontSize >= 60 { // Title
							resEl.Y = -float64(res.Height) * 0.05
							resEl.FontSize = 48
						} else if resEl.FontSize <= 45 { // Subtitle
							resEl.Y = float64(res.Height) * 0.12
							resEl.FontSize = 28
						}
					}
				case "quote":
					if resEl.Type == "rect" && resEl.Width == 1920 && resEl.Height == 1080 {
						resEl.Width = float64(res.Width)
						resEl.Height = float64(res.Height)
					} else if resEl.Type == "text" {
						resEl.X = 0
						resEl.TextAlign = "center"
						if resEl.FontSize >= 40 { // Quote
							resEl.Y = -float64(res.Height) * 0.08
							resEl.FontSize = 36
						} else { // Author
							resEl.Y = float64(res.Height) * 0.15
							resEl.FontSize = 24
						}
					}
				case "outro":
					if resEl.Type == "rect" && resEl.Width == 1920 && resEl.Height == 1080 {
						resEl.Width = float64(res.Width)
						resEl.Height = float64(res.Height)
					} else if resEl.Type == "image" { // Logo
						resEl.X = 0
						resEl.Y = -float64(res.Height) * 0.15
						resEl.Width = float64(res.Width) * 0.4
						resEl.Height = float64(res.Width) * 0.4
					} else if resEl.Type == "text" {
						resEl.X = 0
						resEl.TextAlign = "center"
						if resEl.FontSize >= 50 { // Title
							resEl.Y = float64(res.Height) * 0.12
							resEl.FontSize = 42
						} else if resEl.FontSize <= 38 { // Subtitle
							resEl.Y = float64(res.Height) * 0.22
							resEl.FontSize = 28
						}
					}
				case "image_text":
					if resEl.Type == "rect" && resEl.Width == 1920 && resEl.Height == 1080 {
						resEl.Width = float64(res.Width)
						resEl.Height = float64(res.Height)
					} else if resEl.Type == "image" {
						resEl.X = 0
						resEl.Y = -float64(res.Height) * 0.18
						resEl.Width = float64(res.Width) * 0.65
						resEl.Height = float64(res.Width) * 0.65
					} else if resEl.Type == "text" {
						resEl.X = 0
						resEl.TextAlign = "center"
						if resEl.FontSize >= 50 { // Title (56 or 60)
							resEl.Y = float64(res.Height) * 0.15
							resEl.FontSize = 42
						} else if resEl.FontSize <= 38 { // Paragraph text (32 or 36)
							resEl.Y = float64(res.Height) * 0.28
							resEl.FontSize = 24
						}
					}
				}
			} else {
				// Ajustes para modo paisagem (landscape) nos templates
				if resEl.Type == "rect" && resEl.Width == 1920 && resEl.Height == 1080 {
					resEl.Width = float64(res.Width)
					resEl.Height = float64(res.Height)
				} else {
					resEl.X = resEl.X * scaleX
					resEl.Y = resEl.Y * scaleY
					resEl.Width = resEl.Width * scaleX
					resEl.Height = resEl.Height * scaleY
				}
			}
		} else {
			// Elementos customizados (sem template)
			// Detecta se usa coordenadas absolutas e converte para relativas ao centro
			if resEl.X > 0 && resEl.X <= float64(res.Width) && resEl.Y > 0 && resEl.Y <= float64(res.Height) {
				resEl.X = resEl.X - centerX
				resEl.Y = resEl.Y - centerY
			} else {
				// Se já forem relativas, aplica o escalamento proporcional
				resEl.X = resEl.X * scaleX
				resEl.Y = resEl.Y * scaleY
			}

			if resEl.Type == "rect" && resEl.Width == 1920 && resEl.Height == 1080 {
				resEl.Width = float64(res.Width)
				resEl.Height = float64(res.Height)
			} else {
				resEl.Width = resEl.Width * scaleX
				resEl.Height = resEl.Height * scaleY
			}
		}

		if resEl.Rotation != 0 {
			dc.RotateAbout(gg.Radians(resEl.Rotation), resEl.X, resEl.Y)
		}

		if resEl.Type == "text" {
			dc.SetHexColor(resEl.Color)

			// Escalamento responsivo do tamanho do texto baseando-se no refWidth
			scale := float64(res.Width) / refWidth
			scaledFontSize := resEl.FontSize * scale
			if scaledFontSize < 10 {
				scaledFontSize = 10 // Limite mínimo de legibilidade
			}

			// Utiliza a fonte Roboto pré-carregada para evitar leitura de I/O por frame
			if robotoFont != nil {
				face := truetype.NewFace(robotoFont, &truetype.Options{
					Size:    scaledFontSize,
					DPI:     72,
					Hinting: font.HintingFull,
				})
				dc.SetFontFace(face)
				defer face.Close()
			} else {
				// Fallback se a fonte não foi carregada no cache
				if err := dc.LoadFontFace("assets/fonts/Roboto.ttf", scaledFontSize); err != nil {
					// Fallback silencioso se não houver fonte (graceful degradation)
				}
			}

			align := gg.AlignCenter
			ax, ay := 0.5, 0.5
			if resEl.TextAlign == "left" {
				align = gg.AlignLeft
				ax = 0.0
			} else if resEl.TextAlign == "right" {
				align = gg.AlignRight
				ax = 1.0
			}

			// Calcula o maxWidth responsivo com base no alinhamento e posição X para evitar estouro
			padding := float64(res.Width) * 0.05
			var maxWidth float64

			absX := resEl.X
			if absX < 0 {
				absX = -absX
			}

			if align == gg.AlignLeft {
				maxWidth = (float64(res.Width) / 2.0) - resEl.X - padding
			} else if align == gg.AlignRight {
				maxWidth = (float64(res.Width) / 2.0) + resEl.X - padding
			} else {
				distToEdge := (float64(res.Width) / 2.0) - absX
				maxWidth = 2.0*distToEdge - padding
			}

			minWidth := float64(res.Width) * 0.3
			if maxWidth < minWidth {
				maxWidth = minWidth
			}

			if resEl.ShadowColor != "" {
				// Drop shadow (hard shadow based on offset)
				dc.SetHexColor(resEl.ShadowColor)
				dc.DrawStringWrapped(resEl.Content, resEl.X+resEl.ShadowOffsetX, resEl.Y+resEl.ShadowOffsetY, ax, ay, maxWidth, 1.5, align)
				// Revert color for main text
				dc.SetHexColor(resEl.Color)
			}
			dc.DrawStringWrapped(resEl.Content, resEl.X, resEl.Y, ax, ay, maxWidth, 1.5, align)
		} else if resEl.Type == "rect" {
			// Suporte a renderização de retângulos/banners com gradientes 3D
			if strings.HasPrefix(resEl.Color, "gradient:") {
				colors := strings.Split(strings.TrimPrefix(resEl.Color, "gradient:"), ",")
				if len(colors) == 2 {
					grad := gg.NewLinearGradient(resEl.X, resEl.Y-resEl.Height/2, resEl.X, resEl.Y+resEl.Height/2)
					c1 := parseHexColor(colors[0])
					c2 := parseHexColor(colors[1])
					grad.AddColorStop(0, c1)
					grad.AddColorStop(1, c2)
					dc.SetFillStyle(grad)
				}
			} else {
				dc.SetHexColor(resEl.Color)
			}
			dc.DrawRectangle(resEl.X-resEl.Width/2, resEl.Y-resEl.Height/2, resEl.Width, resEl.Height)
			dc.Fill()
		} else if resEl.Type == "polygon" {
			if len(resEl.Points) > 2 {
				dc.SetHexColor(resEl.Color)
				dc.MoveTo(resEl.Points[0][0], resEl.Points[0][1])
				for i := 1; i < len(resEl.Points); i++ {
					dc.LineTo(resEl.Points[i][0], resEl.Points[i][1])
				}
				dc.ClosePath()
				dc.Fill()
			}
		} else if resEl.Type == "circle" {
			dc.SetHexColor(resEl.Color)
			dc.DrawCircle(resEl.X, resEl.Y, resEl.Width)
			dc.Fill()
		} else if resEl.Type == "frame" {
			dc.SetHexColor(resEl.Color)
			strokeWidth := resEl.StrokeWidth
			if strokeWidth <= 0 {
				strokeWidth = 5.0
			}
			dc.SetLineWidth(strokeWidth)
			dc.DrawRectangle(resEl.X-resEl.Width/2, resEl.Y-resEl.Height/2, resEl.Width, resEl.Height)
			dc.Stroke()
		} else if resEl.Type == "image" {
			var img image.Image
			var exists bool
			if img, exists = imageCache[resEl.Content]; !exists {
				var err error
				img, err = gg.LoadImage(resEl.Content)
				if err != nil {
					dc.Pop()
					continue
				}
			}

			dc.Translate(resEl.X, resEl.Y)
			if resEl.Width > 0 && resEl.Height > 0 {
				imgScaleX := resEl.Width / float64(img.Bounds().Dx())
				imgScaleY := resEl.Height / float64(img.Bounds().Dy())
				dc.Scale(imgScaleX, imgScaleY)
			}
			dc.DrawImageAnchored(img, 0, 0, 0.5, 0.5)
		} else if resEl.Type == "video" {
			framePath := fmt.Sprintf("tmp/frames_%s_%d/frame_%04d.jpg", card.ID, elIdx, frameIndex+1)
			if _, err := os.Stat(framePath); err != nil {
				framePath = fmt.Sprintf("tmp/frames_%s_%d/frame_0001.jpg", card.ID, elIdx)
			}

			img, err := gg.LoadImage(framePath)
			if err != nil {
				// Fallback visual
				dc.SetHexColor("#333333")
				dc.DrawRectangle(resEl.X-resEl.Width/2, resEl.Y-resEl.Height/2, resEl.Width, resEl.Height)
				dc.Fill()
				
				dc.SetHexColor("#FFFFFF")
				dc.DrawStringAnchored("VIDEO PLACEHOLDER", resEl.X, resEl.Y, 0.5, 0.5)
				dc.Pop()
				continue
			}

			dc.Translate(resEl.X, resEl.Y)
			if resEl.Width > 0 && resEl.Height > 0 {
				imgScaleX := resEl.Width / float64(img.Bounds().Dx())
				imgScaleY := resEl.Height / float64(img.Bounds().Dy())
				dc.Scale(imgScaleX, imgScaleY)
			}
			dc.DrawImageAnchored(img, 0, 0, 0.5, 0.5)
		}

		dc.Pop()
	}

	// Legendas da Narração (Subtitles)
	if showSubtitles && card.Narration != "" {
		dc.Push()

		// Fonte e tamanho responsivo baseado em Roboto
		refWidth := 1920.0
		scale := float64(res.Width) / refWidth
		scaledFontSize := 32.0 * scale
		if scaledFontSize < 14 {
			scaledFontSize = 14
		}

		if robotoFont != nil {
			face := truetype.NewFace(robotoFont, &truetype.Options{
				Size:    scaledFontSize,
				DPI:     72,
				Hinting: font.HintingFull,
			})
			dc.SetFontFace(face)
			defer face.Close()
		}

		subX := float64(res.Width) / 2.0
		subY := float64(res.Height) * 0.88
		maxWidth := float64(res.Width) * 0.85

		// Desenhar contorno preto (sombra em 4 direções para legibilidade máxima)
		dc.SetHexColor("#000000")
		dc.DrawStringWrapped(card.Narration, subX+2, subY+2, 0.5, 0.5, maxWidth, 1.4, gg.AlignCenter)
		dc.DrawStringWrapped(card.Narration, subX-2, subY+2, 0.5, 0.5, maxWidth, 1.4, gg.AlignCenter)
		dc.DrawStringWrapped(card.Narration, subX+2, subY-2, 0.5, 0.5, maxWidth, 1.4, gg.AlignCenter)
		dc.DrawStringWrapped(card.Narration, subX-2, subY-2, 0.5, 0.5, maxWidth, 1.4, gg.AlignCenter)

		// Desenhar texto principal branco por cima
		dc.SetHexColor("#ffffff")
		dc.DrawStringWrapped(card.Narration, subX, subY, 0.5, 0.5, maxWidth, 1.4, gg.AlignCenter)

		dc.Pop()
	}
}

