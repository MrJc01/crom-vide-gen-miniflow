import React from 'react';

export default function Canvas({ 
  previewUrl, 
  loading, 
  resolution, 
  elements, 
  selectedElementIdx, 
  onSelectElement, 
  onActionStart 
}) {
  const resWidth = resolution ? resolution.width : 1920;
  const resHeight = resolution ? resolution.height : 1080;
  const ratio = resWidth / resHeight;
  const isVertical = ratio < 1;

  const handleStyle = {
    position: 'absolute',
    width: '12px',
    height: '12px',
    backgroundColor: '#fff',
    border: '2px solid #38bdf8',
    borderRadius: '50%',
    zIndex: 102
  };

  return (
    <div className="canvas-area">
      <div className="overlay-grid"></div>
      
      {loading && <div className="loading">Renderizando...</div>}
      
      <div 
        id="preview-img-container"
        className="preview-container" 
        style={{ 
          aspectRatio: ratio, 
          height: isVertical ? '80vh' : 'auto',
          width: isVertical ? 'auto' : '60vw',
          opacity: loading ? 0.7 : 1,
          position: 'relative',
          containerType: 'size' // Enables cqw/cqh units!
        }}
      >
        {previewUrl ? (
          <>
            <img 
              src={previewUrl} 
              alt="Live Preview" 
              className="preview-image" 
              style={{ width: '100%', height: '100%', display: 'block', pointerEvents: 'none' }}
            />
            {/* Interactive Bounding Boxes Overlays */}
            {elements.map((el, idx) => {
              const isSelected = selectedElementIdx === idx;
              
              // For Text, we inject a hidden text node to get EXACT browser computed width/height!
              const isText = el.type === 'text';

              return (
                <div
                  key={idx}
                  onMouseDown={(e) => {
                    e.stopPropagation();
                    onSelectElement(idx);
                    if (onActionStart) onActionStart(e, idx, 'move');
                  }}
                  style={{
                    position: 'absolute',
                    left: `${(el.x / resWidth) * 100}%`,
                    top: `${(el.y / resHeight) * 100}%`,
                    // Texts auto-size to content! Rects use exact widths.
                    width: isText ? 'auto' : `${((el.width || 0) / resWidth) * 100}%`,
                    height: isText ? 'auto' : `${((el.height || 0) / resHeight) * 100}%`,
                    transform: `translate(-50%, -50%) rotate(${el.rotation || 0}deg)`,
                    cursor: 'move',
                    border: isSelected ? '2px solid #38bdf8' : '1px dashed rgba(255,255,255,0.3)',
                    backgroundColor: isSelected ? 'rgba(56, 189, 248, 0.1)' : 'transparent',
                    boxShadow: isSelected ? '0 0 10px rgba(56, 189, 248, 0.5)' : 'none',
                    zIndex: idx, // Preserve natural z-index so higher elements are always selectable!
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    
                    // The secret magic for responsive text bounds!
                    whiteSpace: 'nowrap',
                    color: 'transparent', // We don't want to see the HTML text, just use it for sizing!
                    fontSize: isText ? `${(el.font_size / resWidth) * 100}cqw` : '1rem',
                    fontFamily: 'Inter, sans-serif',
                    lineHeight: 1,
                    padding: isText ? '0.2cqw' : '0' // Slight padding so border isn't touching pixels
                  }}
                >
                  {/* Invisible text node that stretches the div exactly to the text dimensions */}
                  {isText && el.content}

                  {/* Small tag to identify type when selected */}
                  {isSelected && (
                    <span style={{ 
                      position: 'absolute', 
                      top: '-25px', 
                      left: '-2px', 
                      background: '#38bdf8', 
                      color: '#000', 
                      fontSize: '10px', 
                      padding: '2px 6px',
                      borderRadius: '2px',
                      fontWeight: 'bold',
                      zIndex: 101,
                      letterSpacing: 'normal'
                    }}>
                      {el.type}
                    </span>
                  )}

                  {/* Resize and Rotate Handles */}
                  {isSelected && (
                    <>
                      {/* Rotate Handle */}
                      <div 
                        onMouseDown={(e) => { e.stopPropagation(); onActionStart(e, idx, 'rotate'); }}
                        style={{ ...handleStyle, top: '-30px', left: '50%', transform: 'translate(-50%, 0)', cursor: 'crosshair', backgroundColor: '#fbbf24', borderColor: '#d97706' }}
                      />
                      {/* Top connector line for rotate */}
                      <div style={{ position: 'absolute', top: '-20px', left: '50%', width: '2px', height: '20px', backgroundColor: '#38bdf8' }} />

                      {/* Resize Corners */}
                      <div onMouseDown={(e) => { e.stopPropagation(); onActionStart(e, idx, 'resize-tl'); }} style={{ ...handleStyle, top: '-6px', left: '-6px', cursor: 'nwse-resize' }} />
                      <div onMouseDown={(e) => { e.stopPropagation(); onActionStart(e, idx, 'resize-tr'); }} style={{ ...handleStyle, top: '-6px', right: '-6px', cursor: 'nesw-resize' }} />
                      <div onMouseDown={(e) => { e.stopPropagation(); onActionStart(e, idx, 'resize-bl'); }} style={{ ...handleStyle, bottom: '-6px', left: '-6px', cursor: 'nesw-resize' }} />
                      <div onMouseDown={(e) => { e.stopPropagation(); onActionStart(e, idx, 'resize-br'); }} style={{ ...handleStyle, bottom: '-6px', right: '-6px', cursor: 'nwse-resize' }} />
                    </>
                  )}
                </div>
              )
            })}
          </>
        ) : (
          <div style={{ color: '#64748b', display: 'flex', height: '100%', alignItems:'center', justifyContent: 'center'}}>
            Selecione um template para iniciar
          </div>
        )}
      </div>
    </div>
  )
}
