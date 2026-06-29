import { useState, useEffect, useRef } from 'react';

export default function PreRenderModal({ data, onClose, onRender }) {
  const [formData, setFormData] = useState(null);
  const [durationModes, setDurationModes] = useState({}); // { cardIdx: 'auto' | 'manual' }
  const [videoDurations, setVideoDurations] = useState({}); // { cardIdx: { elIdx: ms } }

  useEffect(() => {
    if (data) {
      const initialData = JSON.parse(JSON.stringify(data));
      // Clear global audio and media contents so they start empty
      initialData.audio_url = '';
      initialData.cards.forEach(card => {
        card.elements.forEach(el => {
          if (el.type === 'video' || el.type === 'image') {
            el.content = '';
          }
        });
      });
      setFormData(initialData);
      
      const initialModes = {};
      initialData.cards.forEach((_, idx) => {
        initialModes[idx] = 'auto'; // Default to auto mode
      });
      setDurationModes(initialModes);
    }
  }, [data]);

  // Recalculate auto durations whenever videoDurations or durationModes change
  useEffect(() => {
    if (!formData) return;
    
    let changed = false;
    const newData = { ...formData };
    
    newData.cards.forEach((card, cIdx) => {
      if (durationModes[cIdx] === 'auto') {
        const durationsInCard = videoDurations[cIdx] ? Object.values(videoDurations[cIdx]) : [];
        if (durationsInCard.length > 0) {
          const maxDur = Math.max(...durationsInCard);
          if (card.duration_ms !== maxDur) {
            card.duration_ms = maxDur;
            changed = true;
          }
        }
      }
    });

    if (changed) {
      setFormData(newData);
    }
  }, [videoDurations, durationModes, formData]);

  if (!formData) return null;

  const handleGlobalAudioChange = (val) => {
    setFormData(prev => ({ ...prev, audio_url: val }));
  };

  const handleMediaChange = (cardIdx, elIdx, val) => {
    const newData = { ...formData };
    newData.cards[cardIdx].elements[elIdx].content = val;
    setFormData(newData);
  };

  const handleFileUpload = async (e, isGlobalAudio, cardIdx, elIdx) => {
    const file = e.target.files[0];
    if (!file) return;

    const uploadData = new FormData();
    uploadData.append('file', file);

    try {
      const res = await fetch('http://localhost:8080/api/upload', {
        method: 'POST',
        body: uploadData
      });
      const result = await res.json();
      if (res.ok && result.path) {
        if (isGlobalAudio) {
          handleGlobalAudioChange(result.path);
        } else {
          handleMediaChange(cardIdx, elIdx, result.path);
          
          if (result.duration_ms && result.duration_ms > 0) {
            setVideoDurations(prev => {
              const prevCard = prev[cardIdx] || {};
              return {
                ...prev,
                [cardIdx]: { ...prevCard, [elIdx]: result.duration_ms }
              };
            });
          }
        }
      } else {
        alert('Erro no upload');
      }
    } catch (err) {
      alert('Falha ao enviar arquivo');
    }
  };

  const handleManualDuration = (cardIdx, ms) => {
    const newData = { ...formData };
    newData.cards[cardIdx].duration_ms = parseInt(ms) || 0;
    setFormData(newData);
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    onRender(formData);
  };

  return (
    <div className="modal-overlay" style={overlayStyle}>
      <div className="modal-content" style={contentStyle}>
        <div className="modal-header" style={headerStyle}>
          <h2>Salvar & Renderizar Vídeo</h2>
          <button className="btn-close" onClick={onClose} style={closeBtnStyle}>×</button>
        </div>
        
        <form onSubmit={handleSubmit} style={formStyle}>
          <p style={{ color: 'var(--text-muted)', marginBottom: '1.5rem', fontSize: '0.9rem' }}>
            O template servirá de blueprint. Preencha os campos abaixo com as mídias e textos finais para injetar no vídeo antes de renderizá-lo!
          </p>

          <div className="form-group" style={{ marginBottom: '2rem', padding: '1rem', border: '1px dashed var(--accent)', borderRadius: '8px', backgroundColor: 'rgba(56, 189, 248, 0.05)' }}>
            <label style={{ color: 'var(--accent)', fontWeight: 'bold', display: 'block', marginBottom: '0.5rem' }}>Música de Fundo Global (Áudio)</label>
            <div style={{ display: 'flex', gap: '0.5rem' }}>
              <div style={{ flex: 1, padding: '0.5rem', color: '#38bdf8', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis', alignSelf: 'center', fontSize: '0.9rem' }} title={formData.audio_url}>
                {formData.audio_url ? `🎵 ${formData.audio_url.split('/').pop()}` : ''}
              </div>
              <label className="btn" style={{ cursor: 'pointer', background: '#334155', color: '#fff', textAlign: 'center', width: 'auto', display: 'flex', alignItems: 'center' }}>
                📁 Escolher Arquivo
                <input type="file" accept="audio/*" style={{ display: 'none' }} onChange={(e) => handleFileUpload(e, true, null, null)} />
              </label>
            </div>
          </div>

          <div className="variables-list" style={{ overflowY: 'auto', maxHeight: '50vh', paddingRight: '10px' }}>
            {formData.cards.map((card, cardIdx) => {
              const variables = card.elements.map((el, elIdx) => ({ el, elIdx }))
                .filter(item => item.el.type === 'video' || item.el.type === 'image' || item.el.type === 'text');

              if (variables.length === 0) return null;

              const mode = durationModes[cardIdx];

              return (
                <div key={cardIdx} style={{ marginBottom: '1.5rem', background: 'var(--bg-main)', padding: '1rem', borderRadius: '8px', border: '1px solid var(--border)' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem', borderBottom: '1px solid #334155', paddingBottom: '0.5rem' }}>
                    <h4 style={{ color: '#94a3b8', margin: 0 }}>🎬 Cena {cardIdx + 1}</h4>
                    
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', fontSize: '0.8rem' }}>
                      <span style={{ color: 'var(--text-muted)' }}>Duração:</span>
                      <select 
                        value={mode} 
                        onChange={(e) => setDurationModes(prev => ({...prev, [cardIdx]: e.target.value}))}
                        style={{ padding: '0.2rem', width: 'auto', fontSize: '0.8rem' }}
                      >
                        <option value="auto">Automático (Maior Vídeo)</option>
                        <option value="manual">Manual (Fixo)</option>
                      </select>
                      
                      <input 
                        type="number"
                        step="0.1"
                        disabled={mode === 'auto'}
                        value={card.duration_ms ? (card.duration_ms / 1000).toFixed(1) : 0}
                        onChange={(e) => handleManualDuration(cardIdx, parseFloat(e.target.value) * 1000)}
                        style={{ width: '80px', padding: '0.2rem', fontSize: '0.8rem', opacity: mode === 'auto' ? 0.5 : 1 }}
                        title="Duração em segundos"
                      />
                      <span style={{ color: 'var(--text-muted)' }}>s</span>
                    </div>
                  </div>

                  {variables.map((item, idx) => {
                    const isMedia = item.el.type === 'video' || item.el.type === 'image';
                    const icon = item.el.type === 'video' ? '🎥' : (item.el.type === 'image' ? '🖼️' : '📝');
                    const label = item.el.type === 'video' ? 'Vídeo' : (item.el.type === 'image' ? 'Imagem' : 'Texto');

                    return (
                      <div className="form-group" key={item.elIdx} style={{ marginLeft: '1rem' }}>
                        <label style={{ display: 'block', marginBottom: '0.3rem' }}>{icon} {label} #{idx + 1}</label>
                        <div style={{ display: 'flex', gap: '0.5rem' }}>
                          {item.el.type === 'text' ? (
                            <textarea 
                              placeholder="Digite o texto final (Opcional)..."
                              value={item.el.content || ''} 
                              onChange={e => handleMediaChange(cardIdx, item.elIdx, e.target.value)} 
                              style={{ flex: 1, padding: '0.75rem', borderRadius: '6px', background: 'var(--bg-main)', color: 'var(--text-main)', border: '1px solid var(--border)', fontFamily: 'inherit' }}
                              rows={2}
                            />
                          ) : (
                            <div style={{ flex: 1, padding: '0.5rem', color: '#38bdf8', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis', display: 'flex', alignItems: 'center', fontSize: '0.9rem' }} title={item.el.content}>
                              {item.el.content ? `📎 ${item.el.content.split('/').pop()}` : ''}
                            </div>
                          )}
                          
                          {isMedia && (
                            <label className="btn" style={{ cursor: 'pointer', background: '#334155', color: '#fff', textAlign: 'center', width: 'auto', display: 'flex', alignItems: 'center' }}>
                              📁 Arquivo
                              <input type="file" accept={item.el.type === 'video' ? "video/*" : "image/*"} style={{ display: 'none' }} onChange={(e) => handleFileUpload(e, false, cardIdx, item.elIdx)} />
                            </label>
                          )}
                        </div>
                      </div>
                    );
                  })}
                </div>
              );
            })}
            
            {formData.cards.every(c => !c.elements.some(e => e.type === 'video' || e.type === 'image' || e.type === 'text')) && (
              <p style={{ textAlign: 'center', color: 'var(--text-muted)', margin: '2rem 0' }}>
                Nenhuma variável dinâmica encontrada no template. O vídeo será gerado exatamente como está no editor!
              </p>
            )}
          </div>

          <div className="modal-footer" style={{ marginTop: '1.5rem', display: 'flex', gap: '1rem', justifyContent: 'flex-end', borderTop: '1px solid var(--border)', paddingTop: '1.5rem' }}>
            <button type="button" className="btn" style={{ background: 'transparent', border: '1px solid var(--border)', width: 'auto' }} onClick={onClose}>Cancelar</button>
            <button type="submit" className="btn" style={{ background: 'var(--accent)', color: '#000', width: 'auto' }}>🚀 Gerar Vídeo Definitivo</button>
          </div>
        </form>
      </div>
    </div>
  );
}

// Inline styles to avoid bloating index.css for this specific modal
const overlayStyle = {
  position: 'fixed',
  top: 0, left: 0, right: 0, bottom: 0,
  backgroundColor: 'rgba(15, 23, 42, 0.85)',
  backdropFilter: 'blur(4px)',
  display: 'flex',
  justifyContent: 'center',
  alignItems: 'center',
  zIndex: 1000
};

const contentStyle = {
  backgroundColor: 'var(--bg-panel)',
  width: '100%',
  maxWidth: '700px',
  borderRadius: '12px',
  border: '1px solid var(--border)',
  boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.5)',
  display: 'flex',
  flexDirection: 'column',
  maxHeight: '90vh'
};

const headerStyle = {
  padding: '1.5rem',
  borderBottom: '1px solid var(--border)',
  display: 'flex',
  justifyContent: 'space-between',
  alignItems: 'center'
};

const closeBtnStyle = {
  background: 'none',
  border: 'none',
  color: 'var(--text-muted)',
  fontSize: '1.5rem',
  cursor: 'pointer'
};

const formStyle = {
  padding: '1.5rem',
  display: 'flex',
  flexDirection: 'column',
  overflow: 'hidden'
};
