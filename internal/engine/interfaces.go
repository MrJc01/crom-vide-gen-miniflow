package engine

import (
	"context"
	"videogen/internal/models"
)

// 9. Interface para o Motor de Renderização
type Renderer interface {
	RenderCard(ctx context.Context, card models.Card, res models.Size, fps int, outPath string) error
}
