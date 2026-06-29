import { useEffect, useRef } from 'react';

export default function Inspector({ 
  data, 
  selectedCardIdx = 0,
  selectedElementIdx, 
  onUpdateElement, 
  onUpdateTemplate,
  onSave, 
  onPreRender,
  onAddElement, 
  onRemoveElement, 
  onMoveUp, 
  onMoveDown 
}) {
  if (!data || !data.cards || data.cards.length === 0) return null;
  const elements = data.cards[selectedCardIdx].elements;
  
  const cardRefs = useRef([]);

  useEffect(() => {
    if (selectedElementIdx !== null && cardRefs.current[selectedElementIdx]) {
      cardRefs.current[selectedElementIdx].scrollIntoView({
        behavior: 'smooth',
        block: 'center',
      });
    }
  }, [selectedElementIdx]);

  return (
    <div className="inspector">
      <div className="panel-header" style={{ gap: '10px' }}>
        Inspector
        <div style={{ display: 'flex', gap: '5px' }}>
          <button className="btn" style={{ padding: '0.4rem', width: 'auto', background: '#334155' }} title="Adicionar Elemento" onClick={onAddElement}>
            ➕
          </button>
          <button className="btn" style={{ padding: '0.4rem 1rem', width: 'auto' }} title="Salvar Template" onClick={onSave}>
            💾
          </button>
          <button className="btn" style={{ padding: '0.4rem 1rem', width: 'auto', background: 'var(--accent)', color: '#000' }} title="Salvar e Pre-Renderizar" onClick={onPreRender}>
            🚀 Render
          </button>
        </div>
      </div>
      <div className="panel-content">
        
        {elements.slice().reverse().map((el, reverseIdx) => {
          const idx = elements.length - 1 - reverseIdx;
          const isSelected = selectedElementIdx === idx;
          const isLast = idx === elements.length - 1;
          const isFirst = idx === 0;
          
          return (
            <div 
              className="element-card" 
              key={idx} 
              ref={el => cardRefs.current[idx] = el}
              style={{
                borderColor: isSelected ? 'var(--accent)' : 'var(--border)',
                boxShadow: isSelected ? 'var(--shadow-glow)' : 'none',
                transform: isSelected ? 'scale(1.02)' : 'scale(1)',
                transition: 'all 0.2s'
              }}
            >
              <div className="element-header">
                <span>{el.type.toUpperCase()} <small style={{color:'#64748b'}}>#{idx}</small></span>
                <div style={{ display: 'flex', gap: '4px' }}>
                  <button className="btn" style={{ padding: '2px 6px', background: '#334155', opacity: isLast ? 0.3 : 1 }} onClick={() => onMoveUp(idx)} title="Subir Camada (Trazer para frente)">⬆️</button>
                  <button className="btn" style={{ padding: '2px 6px', background: '#334155', opacity: isFirst ? 0.3 : 1 }} onClick={() => onMoveDown(idx)} title="Descer Camada (Enviar para trás)">⬇️</button>
                  <button className="btn" style={{ padding: '2px 6px', background: '#ef4444' }} onClick={() => onRemoveElement(idx)} title="Excluir">🗑️</button>
                </div>
              </div>
              
              <div className="form-group">
                <label>Type</label>
                <select value={el.type} onChange={e => onUpdateElement(idx, 'type', e.target.value)}>
                  <option value="text">Text</option>
                  <option value="image">Image</option>
                  <option value="video">Video</option>
                  <option value="rect">Rectangle</option>
                  <option value="circle">Circle</option>
                  <option value="polygon">Polygon</option>
                  <option value="frame">Frame Border</option>
                </select>
              </div>

              {(el.type === 'text' || el.type === 'image' || el.type === 'video') && (
                <div className="form-group">
                  <label>{el.type === 'text' ? 'Text' : 'File Path / URL'}</label>
                  <input 
                    type="text" 
                    value={el.content || ''} 
                    onChange={e => onUpdateElement(idx, 'content', e.target.value)} 
                  />
                </div>
              )}
              
              <div className="row">
                <div className="form-group">
                  <label>X (px)</label>
                  <input 
                    type="number" 
                    value={el.x} 
                    onChange={e => onUpdateElement(idx, 'x', e.target.value)} 
                  />
                </div>
                <div className="form-group">
                  <label>Y (px)</label>
                  <input 
                    type="number" 
                    value={el.y} 
                    onChange={e => onUpdateElement(idx, 'y', e.target.value)} 
                  />
                </div>
                <div className="form-group">
                  <label>Rotation (°)</label>
                  <input 
                    type="number" 
                    value={el.rotation || 0} 
                    onChange={e => onUpdateElement(idx, 'rotation', e.target.value)} 
                  />
                </div>
              </div>

              {(el.type === 'rect' || el.type === 'frame' || el.type === 'video' || el.type === 'image' || el.type === 'circle') && (
                <div className="row">
                  <div className="form-group">
                    <label>Width</label>
                    <input 
                      type="number" 
                      value={el.width || 0} 
                      onChange={e => onUpdateElement(idx, 'width', e.target.value)} 
                    />
                  </div>
                  <div className="form-group">
                    <label>Height</label>
                    <input 
                      type="number" 
                      value={el.height || 0} 
                      onChange={e => onUpdateElement(idx, 'height', e.target.value)} 
                    />
                  </div>
                </div>
              )}

              {(el.type === 'text' || el.type === 'rect' || el.type === 'frame' || el.type === 'circle' || el.type === 'polygon') && (
                <div className="form-group">
                  <label>Color</label>
                  <input 
                    type="text" 
                    value={el.color || ''} 
                    onChange={e => onUpdateElement(idx, 'color', e.target.value)} 
                  />
                </div>
              )}

              {el.type === 'text' && (
                <>
                  <div className="row">
                    <div className="form-group">
                      <label>Font Size</label>
                      <input 
                        type="number" 
                        value={el.font_size || 0} 
                        onChange={e => onUpdateElement(idx, 'font_size', e.target.value)} 
                      />
                    </div>
                    <div className="form-group">
                      <label>Align</label>
                      <select value={el.text_align || 'center'} onChange={e => onUpdateElement(idx, 'text_align', e.target.value)}>
                        <option value="center">Center</option>
                        <option value="left">Left</option>
                        <option value="right">Right</option>
                      </select>
                    </div>
                  </div>
                  <div className="form-group">
                    <label>Shadow Color</label>
                    <input 
                      type="text" 
                      value={el.shadow_color || ''} 
                      onChange={e => onUpdateElement(idx, 'shadow_color', e.target.value)} 
                    />
                  </div>
                </>
              )}

            </div>
          )
        })}
        
      </div>
    </div>
  )
}
