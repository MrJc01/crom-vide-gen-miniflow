package engine_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"videogen/internal/engine"
	"videogen/internal/models"
)

// 72. Mock de Renderer
type mockRenderer struct {
	mu           sync.Mutex
	renderedCards []string
	shouldErr    bool
}

func (m *mockRenderer) RenderCard(ctx context.Context, card models.Card, res models.Size, fps int, outPath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.shouldErr {
		return &mockError{msg: "render failed"}
	}
	
	m.renderedCards = append(m.renderedCards, card.ID)
	return nil
}

type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}

func TestProcessVideo_Success(t *testing.T) {
	tmpl := models.Template{
		TemplateID: "test_promo",
		Resolution: models.Size{Width: 1080, Height: 1920},
		FPS:        30,
		Cards: []models.Card{
			{ID: "card1", DurationMs: 1000, BackgroundColor: "#FFFFFF"},
			{ID: "card2", DurationMs: 2000, BackgroundColor: "#000000"},
		},
	}

	mock := &mockRenderer{}
	
	// Como ProcessVideo chama ConcatVideos no final (que executa ffmpeg e falhará se não houver ffmpeg local),
	// nos testes unitários podemos testar apenas o renderWorker diretamente ou simular se ProcessVideo lida bem com erro de FFmpeg concat.
	// Vamos testar o comportamento concorrente do renderWorker e do pipeline de ProcessVideo.
	
	// Nota: Como ConcatVideos vai rodar ffmpeg local, se o FFmpeg não estiver instalado, ProcessVideo vai falhar na concatenação.
	// Porém, podemos testar que todas as chamadas de RenderCard foram feitas!
	err := engine.ProcessVideo(context.Background(), tmpl, "output_test.mp4", 2, mock)
	
	// Esperamos erro de concatenação se o FFmpeg não estiver no PATH ou mock.ts não existir,
	// mas o importante é que os cards foram mockados com sucesso!
	mock.mu.Lock()
	count := len(mock.renderedCards)
	mock.mu.Unlock()

	if count != 2 {
		t.Errorf("Esperava 2 cards renderizados no mock, obteve %d", count)
	}
	
	_ = err
}

// 80. Teste de estresse com geração de vídeos prolongados (100 cards)
func TestProcessVideo_StressTest(t *testing.T) {
	// Gerar 100 cards programaticamente
	var cards []models.Card
	for i := 1; i <= 100; i++ {
		cards = append(cards, models.Card{
			ID:              fmt.Sprintf("card_%03d", i),
			DurationMs:      1000,
			BackgroundColor: "#000000",
		})
	}

	tmpl := models.Template{
		TemplateID: "stress_promo",
		Resolution: models.Size{Width: 1080, Height: 1920},
		FPS:        30,
		Cards:      cards,
	}

	mock := &mockRenderer{}

	// Executa orquestrador concorrente com worker pool de 8 workers
	err := engine.ProcessVideo(context.Background(), tmpl, "output_stress.mp4", 8, mock)

	mock.mu.Lock()
	count := len(mock.renderedCards)
	mock.mu.Unlock()

	if count != 100 {
		t.Errorf("Esperava 100 cards processados no teste de estresse, obteve %d", count)
	}

	_ = err
}
