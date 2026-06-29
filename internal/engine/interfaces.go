package engine

import "videogen/internal/models"

// 9. Interface para o Motor de Renderização
type Renderer interface {
	RenderCard(card models.Card, res models.Size, fps int, outPath string) error
}
