package engine

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"videogen/internal/models"
)

// ExpandTemplates percorre os cards e gera os elementos correspondentes aos templates selecionados
func ExpandTemplates(tmpl *models.Template) {
	for i := range tmpl.Cards {
		card := &tmpl.Cards[i]
		if card.TemplateName == "" {
			continue
		}
		
		var templateElements []models.Element
		
		switch card.TemplateName {
		case "intro":
			videoPath := card.Parameters["video"]
			imagePath := card.Parameters["image"]
			title := card.Parameters["title"]
			subtitle := card.Parameters["subtitle"]
			
			// 1. Background video or image
			hasBackground := false
			if videoPath != "" {
				templateElements = append(templateElements, models.Element{
					Type:    "video",
					Content: videoPath,
					X:       0,
					Y:       0,
					Width:   1920,
					Height:  1080,
				})
				hasBackground = true
			} else if imagePath != "" {
				templateElements = append(templateElements, models.Element{
					Type:    "image",
					Content: imagePath,
					X:       0,
					Y:       0,
					Width:   1920,
					Height:  1080,
				})
				hasBackground = true
			}
			
			// As requested, if there is a background media, do NOT place anything in front of it.
			// Only draw overlays and text if there is no background media.
			if !hasBackground {
				// 2. Dark background gradient
				templateElements = append(templateElements, models.Element{
					Type:   "rect",
					Color:  "gradient:#0b0b0f,#1a1a2e",
					X:      0,
					Y:      0,
					Width:  1920,
					Height: 1080,
				})
				
				// 3. Title text with shadows
				if title != "" {
					templateElements = append(templateElements, models.Element{
						Type:          "text",
						Content:       title,
						FontSize:      72,
						Color:         "#ffffff",
						X:             0,
						Y:             -100,
						TextAlign:     "center",
						ShadowColor:   "#000000aa",
						ShadowOffsetX: 3,
						ShadowOffsetY: 3,
					})
				}
				
				// 4. Subtitle text with shadows
				if subtitle != "" {
					templateElements = append(templateElements, models.Element{
						Type:          "text",
						Content:       subtitle,
						FontSize:      40,
						Color:         "#00e5ff", // premium neon cyan
						X:             0,
						Y:             100,
						TextAlign:     "center",
						ShadowColor:   "#000000aa",
						ShadowOffsetX: 2,
						ShadowOffsetY: 2,
					})
				}
			}
			
		case "image_text":
			imagePath := card.Parameters["image"]
			title := card.Parameters["title"]
			text := card.Parameters["text"]
			
			// 1. Background gradient rect (space tech deep blue to purple/black)
			templateElements = append(templateElements, models.Element{
				Type:   "rect",
				Color:  "gradient:#0d0d1a,#180a2b",
				X:      0,
				Y:      0,
				Width:  1920,
				Height: 1080,
			})
			
			if imagePath != "" {
				// 2. Translucent glassmorphic card behind the image
				templateElements = append(templateElements, models.Element{
					Type:   "rect",
					Color:  "#ffffff0d", // 5% opacity white
					X:      -450,
					Y:      0,
					Width:  640,
					Height: 640,
				})
				
				// 3. Thin frame border for the card
				templateElements = append(templateElements, models.Element{
					Type:        "frame",
					Color:       "#ffffff1a", // 10% opacity white
					X:           -450,
					Y:           0,
					Width:       640,
					Height:      640,
					StrokeWidth: 2,
				})
				
				// 4. Actual Image centered inside the card
				templateElements = append(templateElements, models.Element{
					Type:    "image",
					Content: imagePath,
					X:       -450,
					Y:       0,
					Width:   600,
					Height:  600,
				})
			}
			
			// 5. Title on the right (cyan neon with shadow)
			if title != "" {
				templateElements = append(templateElements, models.Element{
					Type:          "text",
					Content:       title,
					FontSize:      56,
					Color:         "#00e5ff",
					X:             180,
					Y:             -160,
					TextAlign:     "left",
					ShadowColor:   "#000000aa",
					ShadowOffsetX: 3,
					ShadowOffsetY: 3,
				})
				
				// 6. Horizontal decorative line starting aligned with the title
				templateElements = append(templateElements, models.Element{
					Type:   "rect",
					Color:  "#00e5ff",
					X:      240, // 180 + (120/2) to align left edge at X=180
					Y:      -90,
					Width:  120,
					Height: 4,
				})
			}
			
			// 7. Paragraph text on the right (soft ice blue with shadow)
			if text != "" {
				templateElements = append(templateElements, models.Element{
					Type:          "text",
					Content:       text,
					FontSize:      32,
					Color:         "#e0e0ff",
					X:             180,
					Y:             40,
					TextAlign:     "left",
					ShadowColor:   "#000000aa",
					ShadowOffsetX: 2,
					ShadowOffsetY: 2,
				})
			}
			
		case "quote":
			quote := card.Parameters["quote"]
			author := card.Parameters["author"]
			
			// 1. Background gradient rect (cinematic dark purple to deep blue)
			templateElements = append(templateElements, models.Element{
				Type:   "rect",
				Color:  "gradient:#120c24,#080f1e",
				X:      0,
				Y:      0,
				Width:  1920,
				Height: 1080,
			})
			
			// 2. Huge, ultra-subtle background quotation marks
			templateElements = append(templateElements, models.Element{
				Type:      "text",
				Content:   "“",
				FontSize:  240,
				Color:     "#ffffff08", // 3% opacity
				X:         -600,
				Y:         -200,
				TextAlign: "center",
			})
			templateElements = append(templateElements, models.Element{
				Type:      "text",
				Content:   "”",
				FontSize:  240,
				Color:     "#ffffff08", // 3% opacity
				X:         600,
				Y:         200,
				TextAlign: "center",
			})
			
			// 3. Glowing thin outline frame around the quote card
			templateElements = append(templateElements, models.Element{
				Type:        "frame",
				Color:       "#00e5ff1a", // 10% opacity cyan
				X:           0,
				Y:           0,
				Width:       1500,
				Height:      600,
				StrokeWidth: 2,
			})
			
			// 4. Quote text with strong drop-shadow
			if quote != "" {
				templateElements = append(templateElements, models.Element{
					Type:          "text",
					Content:       fmt.Sprintf("“%s”", quote),
					FontSize:      44,
					Color:         "#ffffff",
					X:             0,
					Y:             -50,
					TextAlign:     "center",
					ShadowColor:   "#000000cc",
					ShadowOffsetX: 3,
					ShadowOffsetY: 3,
				})
			}
			
			// 5. Delicate divider line
			templateElements = append(templateElements, models.Element{
				Type:   "rect",
				Color:  "#00e5ff",
				X:      0,
				Y:      60,
				Width:  80,
				Height: 3,
			})
			
			// 6. Author text (cyan with shadow)
			if author != "" {
				templateElements = append(templateElements, models.Element{
					Type:          "text",
					Content:       author,
					FontSize:      32,
					Color:         "#00e5ff",
					X:             0,
					Y:             130,
					TextAlign:     "center",
					ShadowColor:   "#000000cc",
					ShadowOffsetX: 2,
					ShadowOffsetY: 2,
				})
			}
			
		case "outro":
			logoPath := card.Parameters["logo"]
			title := card.Parameters["title"]
			subtitle := card.Parameters["subtitle"]
			
			// 1. Background gradient rect
			templateElements = append(templateElements, models.Element{
				Type:   "rect",
				Color:  "gradient:#080f1e,#120c24",
				X:      0,
				Y:      0,
				Width:  1920,
				Height: 1080,
			})
			
			// 2. Logo backing card (glassmorphism)
			if logoPath != "" {
				templateElements = append(templateElements, models.Element{
					Type:   "rect",
					Color:  "#ffffff0d", // 5% opacity
					X:      0,
					Y:      -160,
					Width:  320,
					Height: 320,
				})
				
				templateElements = append(templateElements, models.Element{
					Type:        "frame",
					Color:       "#ffffff1a", // 10% opacity
					X:           0,
					Y:           -160,
					Width:       320,
					Height:      320,
					StrokeWidth: 2,
				})
				
				// Actual Logo Image
				templateElements = append(templateElements, models.Element{
					Type:    "image",
					Content: logoPath,
					X:       0,
					Y:       -160,
					Width:   260,
					Height:  260,
				})
			}
			
			// 3. Title text with shadows
			if title != "" {
				templateElements = append(templateElements, models.Element{
					Type:          "text",
					Content:       title,
					FontSize:      52,
					Color:         "#ffffff",
					X:             0,
					Y:             120,
					TextAlign:     "center",
					ShadowColor:   "#000000cc",
					ShadowOffsetX: 3,
					ShadowOffsetY: 3,
				})
				
				// 4. Horizontal decorative line
				templateElements = append(templateElements, models.Element{
					Type:   "rect",
					Color:  "#00e5ff",
					X:      0,
					Y:      180,
					Width:  150,
					Height: 3,
				})
			}
			
			// 5. Subtitle text with shadows
			if subtitle != "" {
				templateElements = append(templateElements, models.Element{
					Type:          "text",
					Content:       subtitle,
					FontSize:      32,
					Color:         "#00e5ff",
					X:             0,
					Y:             240,
					TextAlign:     "center",
					ShadowColor:   "#000000cc",
					ShadowOffsetX: 2,
					ShadowOffsetY: 2,
				})
			}
		}
		
		// Prepend templateElements to card.Elements
		card.Elements = append(templateElements, card.Elements...)
	}
}

// ResolveRelativePaths converte caminhos de arquivos relativos no template em caminhos absolutos baseados no diretório do workspace
func ResolveRelativePaths(tmpl *models.Template, workspaceDir string) {
	if tmpl.AudioURL != "" && !filepath.IsAbs(tmpl.AudioURL) && !strings.HasPrefix(tmpl.AudioURL, "http://") && !strings.HasPrefix(tmpl.AudioURL, "https://") {
		tmpl.AudioURL = filepath.Clean(filepath.Join(workspaceDir, tmpl.AudioURL))
		slog.Info("Caminho da trilha sonora resolvido para absoluto", "path", tmpl.AudioURL)
	}
	for i := range tmpl.Cards {
		card := &tmpl.Cards[i]
		for j := range card.Elements {
			el := &card.Elements[j]
			if (el.Type == "image" || el.Type == "video") && el.Content != "" && !filepath.IsAbs(el.Content) && !strings.HasPrefix(el.Content, "http://") && !strings.HasPrefix(el.Content, "https://") {
				el.Content = filepath.Clean(filepath.Join(workspaceDir, el.Content))
				slog.Info("Caminho do elemento de mídia resolvido para absoluto", "card", card.ID, "type", el.Type, "path", el.Content)
			}
		}
	}
}

// ResolveNarrationAndDurations resolve o TTS e calcula as durações dos cards do template
func ResolveNarrationAndDurations(ctx context.Context, tmpl *models.Template) error {
	for i := range tmpl.Cards {
		card := &tmpl.Cards[i]
		
		// 1. Se houver narração, SEMPRE gera o TTS (independente do modo de cálculo da duração)
		if card.Narration != "" {
			lang := "pt"
			if card.Voice != nil && card.Voice.Lang != "" {
				lang = card.Voice.Lang
			}
			
			ttsPath := filepath.Join("tmp", fmt.Sprintf("tts_%s.mp3", card.ID))
			err := GenerateCromyvoiceTTS(card.Narration, lang, ttsPath)
			if err != nil {
				return fmt.Errorf("erro ao gerar TTS para o card %s: %w", card.ID, err)
			}

			// O card precisa ser definido o tamanho com o tamanho do áudio com 1 segundo de saída e começo (entrada)
			dur, err := GetAudioDuration(ttsPath)
			if err != nil {
				slog.Warn("Falha ao obter duração do TTS com ffprobe, usando fallback", "card", card.ID, "erro", err)
				dur = float64(len(card.Narration)) * 0.1 // Fallback aproximado
			}
			
			// Arredonda para cima para o próximo segundo inteiro (Ceil)
			durSecs := int(dur + 2.0)
			if float64(durSecs) < dur + 2.0 {
				durSecs++
			}
			card.DurationMs = durSecs * 1000
			slog.Info("Duração do card ajustada dinamicamente via TTS (TTS + 2.0s Ceil)", "card", card.ID, "duracao_ms", card.DurationMs)
		}
		
		// 2. Determina o modo de duração padrão se não estiver definido
		mode := card.DurationMode
		if mode == "" {
			if card.DurationMs > 0 {
				mode = "manual"
			} else if card.Narration != "" {
				mode = "narration"
			} else if hasVideoElement(card) {
				mode = "video"
			} else {
				mode = "manual"
			}
		}
		
		switch mode {
		case "narration":
			if card.DurationMs <= 0 {
				card.DurationMs = 5000 // Fallback se não tiver narração mas o modo for narration
			}
			
		case "video":
			// Encontra o maior vídeo
			maxDur := 0.0
			for _, el := range card.Elements {
				if el.Type == "video" && el.Content != "" {
					dur, err := GetAudioDuration(el.Content) // Funciona para vídeo também
					if err == nil && dur > maxDur {
						maxDur = dur
					}
				}
			}
			
			if maxDur > 0 {
				card.DurationMs = int(maxDur * 1000)
				slog.Info("Duração do card calculada via maior vídeo", "card", card.ID, "duracao_ms", card.DurationMs)
			} else {
				slog.Warn("Card configurado para modo 'video' mas nenhum vídeo válido foi encontrado", "card", card.ID)
				if card.DurationMs <= 0 {
					card.DurationMs = 5000 // Fallback
				}
			}
			
		case "manual":
			if card.DurationMs <= 0 {
				card.DurationMs = 5000 // Fallback
			}
		}
	}
	return nil
}

func hasVideoElement(card *models.Card) bool {
	for _, el := range card.Elements {
		if el.Type == "video" {
			return true
		}
	}
	return false
}

// GenerateCromyvoiceTTS tenta usar o cromyvoice para gerar a narração, com fallback para o Google TTS.
func GenerateCromyvoiceTTS(text, lang, outPath string) error {
	binaryPath := "./bin/cromyvoice"
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		binaryPath = "./cromyvoice"
		if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
			slog.Info("Binário cromyvoice não encontrado, usando fallback Google TTS")
			return GenerateGoogleTTS(text, lang, outPath)
		}
	}

	slog.Info("Gerando TTS via cromyvoice", "text", text, "path", outPath)
	
	args := []string{"-text", text, "-out", outPath}
	if lang != "" && lang != "pt" {
		if lang == "en" {
			args = append(args, "-voice", "en-US-GuyNeural")
		} else if lang == "es" {
			args = append(args, "-voice", "es-ES-AlvaroNeural")
		}
	}

	cmd := exec.Command(binaryPath, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		slog.Warn("Falha ao gerar TTS com cromyvoice, usando fallback Google TTS", "erro", err, "output", string(out))
		return GenerateGoogleTTS(text, lang, outPath)
	}

	return nil
}

// GenerateGoogleTTS gera áudio de TTS dividindo o texto em partes de 200 caracteres
func GenerateGoogleTTS(text, lang, outPath string) error {
	// Divide o texto por pontuação/espaço em partes de no máximo 200 caracteres
	parts := splitText(text, 200)
	
	var tempFiles []string
	defer func() {
		for _, f := range tempFiles {
			_ = os.Remove(f)
		}
	}()
	
	for idx, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		
		tempFile := fmt.Sprintf("%s.part%d.mp3", outPath, idx)
		err := downloadTTSChunk(part, lang, tempFile)
		if err != nil {
			return err
		}
		tempFiles = append(tempFiles, tempFile)
	}
	
	// Concatena os arquivos MP3 simplesmente juntando seus bytes
	outFile, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("erro ao criar arquivo de saída do TTS: %w", err)
	}
	defer outFile.Close()
	
	for _, tempFile := range tempFiles {
		f, err := os.Open(tempFile)
		if err != nil {
			return fmt.Errorf("erro ao abrir parte temporária do TTS: %w", err)
		}
		_, err = io.Copy(outFile, f)
		f.Close()
		if err != nil {
			return fmt.Errorf("erro ao concatenar parte do TTS: %w", err)
		}
	}
	
	return nil
}

func downloadTTSChunk(text, lang, filepath string) error {
	baseURL := "https://translate.google.com/translate_tts"
	u, err := url.Parse(baseURL)
	if err != nil {
		return err
	}
	
	q := u.Query()
	q.Set("ie", "UTF-8")
	q.Set("tl", lang)
	q.Set("client", "tw-ob")
	q.Set("q", text)
	u.RawQuery = q.Encode()
	
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("erro ao fazer requisição de TTS: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status de resposta do TTS inválido: %d", resp.StatusCode)
	}
	
	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("erro ao criar arquivo de parte do TTS: %w", err)
	}
	defer out.Close()
	
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("erro ao salvar parte do TTS: %w", err)
	}
	
	return nil
}

func GetAudioDuration(path string) (float64, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", path)
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	
	durStr := strings.TrimSpace(string(out))
	dur, err := strconv.ParseFloat(durStr, 64)
	if err != nil {
		return 0, err
	}
	
	return dur, nil
}

// splitText divide o texto em pedaços de no máximo maxLen caracteres, sem cortar palavras no meio
func splitText(text string, maxLen int) []string {
	var chunks []string
	words := strings.Fields(text)
	
	var currentChunk strings.Builder
	for _, word := range words {
		if currentChunk.Len()+len(word)+1 > maxLen {
			chunks = append(chunks, currentChunk.String())
			currentChunk.Reset()
		}
		if currentChunk.Len() > 0 {
			currentChunk.WriteString(" ")
		}
		currentChunk.WriteString(word)
	}
	if currentChunk.Len() > 0 {
		chunks = append(chunks, currentChunk.String())
	}
	
	return chunks
}
