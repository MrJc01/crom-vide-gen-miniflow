export default function Timeline({ cards, selectedCardIdx, onSelectCard, onAddCard, onRemoveCard, onUpdateDuration }) {
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
              <span className="card-number">{idx + 1}</span>
              {cards.length > 1 && (
                <button className="delete-card-btn" onClick={(e) => { e.stopPropagation(); onRemoveCard(idx); }}>×</button>
              )}
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
