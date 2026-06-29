export default function Timeline({ cards, selectedCardIdx, onSelectCard, onAddCard, onRemoveCard, onCloneCard, onMoveCard, onUpdateDuration }) {
  if (!cards) return null;

  return (
    <div className="timeline-area">
      <div className="timeline-header">
        <span>Cenas (Cards)</span>
        <button className="btn" style={{ padding: '0.2rem 0.5rem', width: 'auto', fontSize: '0.8rem' }} onClick={onAddCard}>➕ Cena</button>
      </div>
      
      <div className="timeline-scroll">
        {cards.map((card, idx) => (
          <div 
            key={card.id || idx} 
            className={`timeline-card ${selectedCardIdx === idx ? 'active' : ''}`}
            onClick={() => onSelectCard(idx)}
          >
            <div className="timeline-card-header">
              <span className="card-number">#{idx + 1}</span>
              <div className="timeline-card-actions" style={{ display: 'flex', gap: '0.2rem' }}>
                <button 
                  className="timeline-action-btn" 
                  title="Clonar Cena" 
                  onClick={(e) => { e.stopPropagation(); onCloneCard(idx); }}
                >
                  ❐
                </button>
                <button 
                  className="timeline-action-btn" 
                  title="Mover para Esquerda" 
                  disabled={idx === 0}
                  onClick={(e) => { e.stopPropagation(); onMoveCard(idx, -1); }}
                >
                  ←
                </button>
                <button 
                  className="timeline-action-btn" 
                  title="Mover para Direita" 
                  disabled={idx === cards.length - 1}
                  onClick={(e) => { e.stopPropagation(); onMoveCard(idx, 1); }}
                >
                  →
                </button>
                {cards.length > 1 && (
                  <button 
                    className="timeline-action-btn delete" 
                    title="Excluir Cena" 
                    onClick={(e) => { e.stopPropagation(); onRemoveCard(idx); }}
                  >
                    ×
                  </button>
                )}
              </div>
            </div>
            <div className="timeline-card-preview" style={{ backgroundColor: card.background_color || '#1e293b' }}>
              {/* Simple representation of elements */}
              <span style={{ fontSize: '10px', color: '#fff', opacity: 0.5 }}>{card.elements?.length || 0} items</span>
            </div>
            <div className="timeline-card-footer" onClick={(e) => e.stopPropagation()}>
              <input 
                type="number" 
                className="duration-input"
                value={card.duration_ms || 0} 
                onChange={(e) => onUpdateDuration(idx, e.target.value)}
                title="Duração em ms"
              />
              <span style={{ fontSize: '10px', color: '#64748b' }}>ms</span>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
