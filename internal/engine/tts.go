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
			if videoPath != "" {
				templateElements = append(templateElements, models.Element{
					Type:    "video",
					Content: videoPath,
					X:       0,
					Y:       0,
					Width:   1920,
					Height:  1080,
				})
			} else if imagePath != "" {
				templateElements = append(templateElements, models.Element{
					Type:    "image",
					Content: imagePath,
					X:       0,
					Y:       0,
					Width:   1920,
					Height:  1080,
				})
			}
			
			// 2. Dark overlay (a semi-transparent rect)
			templateElements = append(templateElements, models.Element{
				Type:   "rect",
				Color:  "#000000cc", // 80% opacity black
				X:      0,
				Y:      0,
				Width:  1920,
				Height: 1080,
			})
			
			// 3. Title text
			if title != "" {
				templateElements = append(templateElements, models.Element{
					Type:      "text",
					Content:   title,
					FontSize:  72,
					Color:     "#ffffff",
					X:         0,
					Y:         -100,
					TextAlign: "center",
				})
			}
			
			// 4. Subtitle text
			if subtitle != "" {
				templateElements = append(templateElements, models.Element{
					Type:      "text",
					Content:   subtitle,
					FontSize:  40,
					Color:     "#00adb5", // nice cyan
					X:         0,
					Y:         100,
					TextAlign: "center",
				})
			}
			
		case "image_text":
			imagePath := card.Parameters["image"]
			title := card.Parameters["title"]
			text := card.Parameters["text"]
			
			// 1. Background gradient rect
			templateElements = append(templateElements, models.Element{
				Type:   "rect",
				Color:  "gradient:#1a1a2e,#0f0f1b",
				X:      0,
				Y:      0,
				Width:  1920,
				Height: 1080,
			})
			
			// 2. Image on the left
			if imagePath != "" {
				templateElements = append(templateElements, models.Element{
					Type:    "image",
					Content: imagePath,
					X:       -400,
					Y:       0,
					Width:   800,
					Height:  800,
				})
			}
			
			// 3. Title on the right
			if title != "" {
				templateElements = append(templateElements, models.Element{
					Type:      "text",
					Content:   title,
					FontSize:  60,
					Color:     "#ffffff",
					X:         150,
					Y:         -150,
					TextAlign: "left",
				})
			}
			
			// 4. Paragraph text on the right
			if text != "" {
				templateElements = append(templateElements, models.Element{
					Type:      "text",
					Content:   text,
					FontSize:  36,
					Color:     "#eeeeee",
					X:         150,
					Y:         100,
					TextAlign: "left",
				})
			}
			
		case "quote":
			quote := card.Parameters["quote"]
			author := card.Parameters["author"]
			
			// 1. Background gradient rect
			templateElements = append(templateElements, models.Element{
				Type:   "rect",
				Color:  "gradient:#0f0f1b,#1a1a2e",
				X:      0,
				Y:      0,
				Width:  1920,
				Height: 1080,
			})
			
			// 2. Quote text
			if quote != "" {
				templateElements = append(templateElements, models.Element{
					Type:      "text",
					Content:   fmt.Sprintf(`"%s"`, quote),
					FontSize:  54,
					Color:     "#eeeeee",
					X:         0,
					Y:         -50,
					TextAlign: "center",
				})
			}
			
			// 3. Author text
			if author != "" {
				templateElements = append(templateElements, models.Element{
					Type:      "text",
					Content:   fmt.Sprintf("- %s", author),
					FontSize:  36,
					Color:     "#00adb5",
					X:         0,
					Y:         150,
					TextAlign: "center",
				})
			}
			
		case "outro":
			logoPath := card.Parameters["logo"]
			title := card.Parameters["title"]
			subtitle := card.Parameters["subtitle"]
			
			// 1. Background gradient rect
			templateElements = append(templateElements, models.Element{
				Type:   "rect",
				Color:  "gradient:#1a1a2e,#0f0f1b",
				X:      0,
				Y:      0,
				Width:  1920,
				Height: 1080,
			})
			
			// 2. Logo image
			if logoPath != "" {
				templateElements = append(templateElements, models.Element{
					Type:    "image",
					Content: logoPath,
					X:       0,
					Y:       -150,
					Width:   350,
					Height:  350,
				})
			}
			
			// 3. Title text
			if title != "" {
				templateElements = append(templateElements, models.Element{
					Type:      "text",
					Content:   title,
					FontSize:  54,
					Color:     "#ffffff",
					X:         0,
					Y:         150,
					TextAlign: "center",
				})
			}
			
			// 4. Subtitle text
			if subtitle != "" {
				templateElements = append(templateElements, models.Element{
					Type:      "text",
					Content:   subtitle,
					FontSize:  36,
					Color:     "#eeeeee",
					X:         0,
					Y:         250,
					TextAlign: "center",
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
		
		// Determina o modo de duração padrão se não estiver definido
		mode := card.DurationMode
		if mode == "" {
			if card.Narration != "" {
				mode = "narration"
			} else if hasVideoElement(card) {
				mode = "video"
			} else {
				mode = "manual"
			}
		}
		
		switch mode {
		case "narration":
			if card.Narration == "" {
				slog.Warn("Card configurado para modo 'narration' mas não possui texto de narração", "card", card.ID)
				if card.DurationMs <= 0 {
					card.DurationMs = 5000 // Fallback
				}
				continue
			}
			
			// Determina o idioma do TTS
			lang := "pt"
			if card.Voice != nil && card.Voice.Lang != "" {
				lang = card.Voice.Lang
			}
			
			// Gera o arquivo de TTS
			ttsPath := filepath.Join("tmp", fmt.Sprintf("tts_%s.mp3", card.ID))
			err := GenerateCromyvoiceTTS(card.Narration, lang, ttsPath)
			if err != nil {
				return fmt.Errorf("erro ao gerar TTS para o card %s: %w", card.ID, err)
			}
			
			// Mede a duração do TTS com ffprobe
			dur, err := GetAudioDuration(ttsPath)
			if err != nil {
				slog.Warn("Falha ao obter duração do TTS com ffprobe, usando fallback", "card", card.ID, "erro", err)
				dur = float64(len(card.Narration)) * 0.1 // Fallback aproximado
			}
			
			// Define a duração: tempo do áudio + 1500ms de margem (entre 1 e 2 segundos)
			card.DurationMs = int(dur * 1000) + 1500
			slog.Info("Duração do card calculada via TTS", "card", card.ID, "duracao_ms", card.DurationMs)
			
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
