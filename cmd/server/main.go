package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/fogleman/gg"
	"videogen/internal/assets"
	"videogen/internal/db"
	"videogen/internal/engine"
	"videogen/internal/models"
)

const templatesDir = "templates/examples"

func main() {
	// Extrair recursos embutidos padrões (bootstrapping)
	if err := assets.ExtractAll(); err != nil {
		log.Printf("Aviso: erro ao extrair recursos embutidos: %v", err)
	}

	// Prepara PATH local para encontrar o ffmpeg
	if absBin, err := filepath.Abs("bin"); err == nil {
		os.Setenv("PATH", absBin+string(os.PathListSeparator)+os.Getenv("PATH"))
	}

	// Inicializar Banco de Dados JSON
	if err := db.InitDB(); err != nil {
		log.Fatalf("Falha ao inicializar DB: %v", err)
	}

	// Configurar rotas CORS e API
	mux := http.NewServeMux()

	mux.HandleFunc("/api/templates", handleTemplates)
	mux.HandleFunc("/api/templates/", handleTemplateByID)
	mux.HandleFunc("/api/preview", handlePreview)
	mux.HandleFunc("/api/render", handleRender)
	mux.HandleFunc("/api/upload", handleUpload)
	
	// API de Vídeos
	mux.HandleFunc("/api/videos", handleVideos)
	mux.HandleFunc("/api/videos/", handleVideoByID)

	// Servir arquivos estáticos dos vídeos gerados
	os.MkdirAll("outputs", 0755)
	mux.Handle("/outputs/", http.StripPrefix("/outputs/", http.FileServer(http.Dir("outputs"))))

	// Iniciar servidor
	port := ":8080"
	log.Printf("Iniciando servidor de API na porta %s...", port)
	log.Fatal(http.ListenAndServe(port, enableCORS(mux)))
}

// Middleware CORS simples para desenvolvimento
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// GET /api/templates - Lista os arquivos JSON
func handleTemplates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	files, err := os.ReadDir(templatesDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var templates []string
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".json") {
			templates = append(templates, strings.TrimSuffix(f.Name(), ".json"))
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(templates)
}

// GET /api/templates/{id} e POST /api/templates/{id}
func handleTemplateByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/templates/")
	if id == "" {
		http.Error(w, "Missing ID", http.StatusBadRequest)
		return
	}

	isExample := false
	if strings.HasSuffix(id, "/example") {
		id = strings.TrimSuffix(id, "/example")
		isExample = true
	}
	
	path := filepath.Join(templatesDir, id+".json")

	if r.Method == http.MethodGet {
		data, err := os.ReadFile(path)
		if err != nil {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		if isExample {
			var tmpl models.Template
			if err := json.Unmarshal(data, &tmpl); err != nil {
				http.Error(w, "Error parsing template: "+err.Error(), http.StatusInternalServerError)
				return
			}
			schemaText := tmpl.GenerateSchemaPrint()

			response := struct {
				Template    models.Template `json:"template"`
				SchemaPrint string          `json:"schema_print"`
			}{
				Template:    tmpl,
				SchemaPrint: schemaText,
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	} else if r.Method == http.MethodPost {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		// Validar JSON
		var tmpl models.Template
		if err := json.Unmarshal(data, &tmpl); err != nil {
			http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Salvar formatado
		formatted, _ := json.MarshalIndent(tmpl, "", "  ")
		if err := os.WriteFile(path, formatted, 0644); err != nil {
			http.Error(w, "Failed to save", http.StatusInternalServerError)
			return
		}
		
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success"}`))
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// POST /api/preview - Recebe o JSON do template inteiro, renderiza o primeiro card, retorna PNG
func handlePreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var tmpl models.Template
	if err := json.NewDecoder(r.Body).Decode(&tmpl); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if len(tmpl.Cards) == 0 {
		http.Error(w, "No cards to preview", http.StatusBadRequest)
		return
	}

	card := tmpl.Cards[0]
	res := tmpl.Resolution
	if res.Width == 0 || res.Height == 0 {
		res = models.Size{Width: 1920, Height: 1080}
	}

	// Desenhar o canvas em memória!
	dc := gg.NewContext(res.Width, res.Height)
	
	// Utilizar nossa engine atualizada!
	engine.DrawCardState(dc, card, res, 0, nil, nil)

	// Codificar imagem final para PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, dc.Image()); err != nil {
		http.Error(w, "Failed to encode image", http.StatusInternalServerError)
		return
	}

	// Retornar a imagem binária
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", buf.Len()))
	w.Write(buf.Bytes())
}

// POST /api/render - Inicia a renderização do vídeo em background
func handleRender(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var tmpl models.Template
	if err := json.NewDecoder(r.Body).Decode(&tmpl); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := tmpl.Validate(); err != nil {
		http.Error(w, "Invalid Template: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Criar nome de output único ou baseado no template
	jobID := fmt.Sprintf("job_%d", time.Now().UnixMilli())
	outPath := filepath.Join("outputs", fmt.Sprintf("output_%s_%s.mp4", tmpl.TemplateID, jobID))

	job := db.VideoJob{
		ID:         jobID,
		TemplateID: tmpl.TemplateID,
		Status:     "rendering",
		FilePath:   outPath,
		Category:   "Uncategorized",
		Archived:   false,
	}
	if err := db.AddVideoJob(job); err != nil {
		log.Printf("Error saving video job: %v", err)
	}

	// Rodar em background (Goroutine)
	go func(t models.Template, output string, jID string) {
		log.Printf("Starting background render for template %s to %s", t.TemplateID, output)
		
		renderer := engine.NewFFmpegRenderer(t.HWAccel, t.JPEGQuality) 
		
		err := engine.ProcessVideo(context.Background(), t, output, runtime.NumCPU(), renderer)
		if err != nil {
			log.Printf("ERROR rendering video %s: %v", t.TemplateID, err)
			db.UpdateVideoJobStatus(jID, "error", err.Error())
		} else {
			log.Printf("SUCCESS rendering video %s: %v", t.TemplateID, output)
			db.UpdateVideoJobStatus(jID, "done", "")
		}
	}(tmpl, outPath, jobID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(fmt.Sprintf(`{"status":"accepted", "message":"Renderização iniciada", "job_id":"%s"}`, jobID)))
}

// POST /api/upload - Recebe um arquivo via multipart/form-data e salva em tmp/uploads/
func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limitar tamanho do upload para 500MB
	err := r.ParseMultipartForm(500 << 20)
	if err != nil {
		http.Error(w, "File too large or invalid form: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Garantir diretório de uploads
	uploadDir := filepath.Join("tmp", "uploads")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		http.Error(w, "Failed to create upload directory", http.StatusInternalServerError)
		return
	}

	// Usar o nome original com cuidado, mas para MVP local tá ok.
	// Idealmente geraria um UUID para evitar colisões
	filename := filepath.Base(header.Filename)
	filePath := filepath.Join(uploadDir, filename)

	out, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		http.Error(w, "Failed to write file", http.StatusInternalServerError)
		return
	}

	// Tenta extrair a duração usando ffprobe (se falhar ou for imagem, retorna 0)
	durationMs := 0
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", filePath)
	outProbe, errProbe := cmd.Output()
	if errProbe == nil {
		durStr := strings.TrimSpace(string(outProbe))
		var durSecs float64
		if _, errParse := fmt.Sscanf(durStr, "%f", &durSecs); errParse == nil {
			durationMs = int(durSecs * 1000)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"status":"success", "path":"%s", "duration_ms":%d}`, filePath, durationMs)))
}

// GET /api/videos - Retorna todos os vídeos
func handleVideos(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jobs, err := db.ReadDB()
	if err != nil {
		http.Error(w, "Failed to read DB: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobs)
}

// GET /api/videos/:id e DELETE /api/videos/:id e PUT /api/videos/:id
func handleVideoByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/videos/")
	if id == "" {
		http.Error(w, "Missing ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		jobs, err := db.ReadDB()
		if err != nil {
			http.Error(w, "Failed to read DB", http.StatusInternalServerError)
			return
		}
		if job, ok := jobs[id]; ok {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(job)
		} else {
			http.Error(w, "Not found", http.StatusNotFound)
		}
	case http.MethodDelete:
		if err := db.DeleteVideoJob(id); err != nil {
			http.Error(w, "Failed to delete", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"deleted"}`))
	case http.MethodPut:
		var req struct {
			Category string `json:"category"`
			Archived bool   `json:"archived"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		if err := db.UpdateVideoJobMeta(id, req.Category, req.Archived); err != nil {
			http.Error(w, "Failed to update", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"updated"}`))
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
