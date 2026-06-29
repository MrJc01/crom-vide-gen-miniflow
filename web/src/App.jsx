import { useState, useEffect, useCallback, useRef } from 'react'
import './index.css'
import Sidebar from './Sidebar'
import Canvas from './Canvas'
import Inspector from './Inspector'
import Timeline from './Timeline'
import PreRenderModal from './PreRenderModal'

const API_BASE = 'http://localhost:8080/api';

function App() {
  const [templates, setTemplates] = useState([]);
  const [selectedId, setSelectedId] = useState(null);
  const [currentData, setCurrentData] = useState(null);
  const [previewUrl, setPreviewUrl] = useState(null);
  const [loading, setLoading] = useState(false);
  const [selectedCardIdx, setSelectedCardIdx] = useState(0);
  const [isPreRenderOpen, setIsPreRenderOpen] = useState(false);
  
  // Selection and Dragging State
  const [selectedElementIdx, setSelectedElementIdx] = useState(null);
  const [isDragging, setIsDragging] = useState(false);
  const dragRef = useRef({ startX: 0, startY: 0, initialElX: 0, initialElY: 0 });

  // Fetch templates list
  const loadTemplates = useCallback(() => {
    fetch(`${API_BASE}/templates`)
      .then(res => res.json())
      .then(data => setTemplates(data || []))
      .catch(err => console.error("Failed to load templates", err));
  }, []);

  useEffect(() => {
    loadTemplates();
  }, [loadTemplates]);

  // Load specific template
  useEffect(() => {
    if (!selectedId) return;
    setLoading(true);
    setSelectedElementIdx(null); // Reset selection
    setSelectedCardIdx(0);
    fetch(`${API_BASE}/templates/${selectedId}`)
      .then(res => res.json())
      .then(data => {
        setCurrentData(data);
        updatePreview(data);
      })
      .catch(err => console.error(err));
  }, [selectedId]);

  const updatePreview = async (data) => {
    setLoading(true);
    try {
      const res = await fetch(`${API_BASE}/preview`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data)
      });
      if (res.ok) {
        const blob = await res.blob();
        setPreviewUrl(URL.createObjectURL(blob));
      }
    } catch (err) {
      console.error(err);
    }
    setLoading(false);
  };

  const handleUpdateElement = (elIdx, field, value) => {
    if (!currentData) return;
    const newData = { ...currentData };
    
    // Auto convert types
    let parsedVal = value;
    if (['x', 'y', 'width', 'height', 'font_size', 'stroke_width', 'rotation'].includes(field)) {
      parsedVal = parseFloat(value) || 0;
    }
    
    newData.cards[selectedCardIdx].elements[elIdx][field] = parsedVal;
    setCurrentData(newData);
    
    // Debounce preview update
    clearTimeout(window.previewTimeout);
    window.previewTimeout = setTimeout(() => {
      updatePreview(newData);
    }, 300);
  };

  // Z-Index and Component Management
  const handleRemoveElement = (idx) => {
    if (!currentData) return;
    const newData = { ...currentData };
    newData.cards[selectedCardIdx].elements.splice(idx, 1);
    setSelectedElementIdx(null);
    setCurrentData(newData);
    updatePreview(newData);
  };

  const handleMoveUp = (idx) => {
    if (!currentData || idx >= currentData.cards[selectedCardIdx].elements.length - 1) return;
    const newData = { ...currentData };
    const elements = newData.cards[selectedCardIdx].elements;
    [elements[idx], elements[idx+1]] = [elements[idx+1], elements[idx]];
    setSelectedElementIdx(idx + 1);
    setCurrentData(newData);
    updatePreview(newData);
  };

  const handleMoveDown = (idx) => {
    if (!currentData || idx <= 0) return;
    const newData = { ...currentData };
    const elements = newData.cards[selectedCardIdx].elements;
    [elements[idx], elements[idx-1]] = [elements[idx-1], elements[idx]];
    setSelectedElementIdx(idx - 1);
    setCurrentData(newData);
    updatePreview(newData);
  };

  const handleAddElement = () => {
    if (!currentData) return;
    const newData = { ...currentData };
    newData.cards[selectedCardIdx].elements.push({
      type: "text",
      content: "NOVO TEXTO",
      font_size: 60,
      color: "#FFFFFF",
      x: 960,
      y: 540,
      text_align: "center"
    });
    setSelectedElementIdx(newData.cards[selectedCardIdx].elements.length - 1);
    setCurrentData(newData);
    updatePreview(newData);
  };

  // Card Management
  const handleAddCard = () => {
    if (!currentData) return;
    const newData = { ...currentData };
    newData.cards.push({
      id: `card_${newData.cards.length + 1}`,
      duration_ms: 3000,
      background_color: "#1e293b",
      elements: []
    });
    setCurrentData(newData);
    setSelectedCardIdx(newData.cards.length - 1);
    setSelectedElementIdx(null);
    updatePreview(newData);
  };

  const handleRemoveCard = (idx) => {
    if (!currentData || currentData.cards.length <= 1) return;
    const newData = { ...currentData };
    newData.cards.splice(idx, 1);
    const newIdx = Math.min(selectedCardIdx, newData.cards.length - 1);
    setCurrentData(newData);
    setSelectedCardIdx(newIdx);
    setSelectedElementIdx(null);
    updatePreview(newData);
  };

  const handleUpdateCardDuration = (idx, value) => {
    if (!currentData) return;
    const newData = { ...currentData };
    newData.cards[idx].duration_ms = parseInt(value) || 0;
    setCurrentData(newData);
    // don't need to update preview just for duration
  };

  const handleUpdateTemplate = (field, value) => {
    if (!currentData) return;
    const newData = { ...currentData };
    if (field === 'fps') {
      newData[field] = parseInt(value) || 30;
    } else {
      newData[field] = value;
    }
    setCurrentData(newData);
  };

  const handleNewTemplate = async (name) => {
    const defaultData = {
      template_id: name,
      resolution: { width: 1920, height: 1080 },
      fps: 30,
      cards: [
        {
          id: "card_1",
          duration_ms: 5000,
          background_color: "#1e293b",
          elements: []
        }
      ]
    };
    try {
      await fetch(`${API_BASE}/templates/${name}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(defaultData)
      });
      loadTemplates();
      setSelectedId(name);
    } catch (err) {
      alert('Erro ao criar template');
    }
  };

  const handleSave = async () => {
    if (!currentData || !selectedId) return;
    try {
      await fetch(`${API_BASE}/templates/${selectedId}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(currentData)
      });
      alert('Template salvo com sucesso!');
    } catch (err) {
      alert('Erro ao salvar template');
    }
  };

  const handleFinalRender = async (finalData) => {
    try {
      // Optamos por salvar antes de renderizar para garantir que os dados de estrutura (tamanho, cor, posições) existam no disco
      // Mas o json finalData enviado aqui tem as URL's injetadas. Vamos enviar direto para o /api/render!
      const res = await fetch(`${API_BASE}/render`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(finalData)
      });
      const data = await res.json();
      if (res.ok) {
        alert(data.message || 'Renderização Iniciada com sucesso!');
        setIsPreRenderOpen(false);
      } else {
        alert('Erro ao iniciar renderização: ' + data.error);
      }
    } catch (err) {
      alert('Erro de conexão ao tentar renderizar.');
    }
  };

  // Drag, Resize and Rotate Logic
  const handleActionStart = (e, idx, action) => {
    e.preventDefault();
    setSelectedElementIdx(idx);
    setIsDragging(true);
    
    const el = currentData.cards[selectedCardIdx].elements[idx];
    const res = currentData.resolution || { width: 1920, height: 1080 };
    
    let centerX = 0;
    let centerY = 0;
    const container = document.getElementById('preview-img-container');
    if (container) {
      const rect = container.getBoundingClientRect();
      const scaleX = rect.width / res.width;
      const scaleY = rect.height / res.height;
      centerX = rect.left + (el.x * scaleX);
      centerY = rect.top + (el.y * scaleY);
    }

    dragRef.current = {
      action,
      startX: e.clientX,
      startY: e.clientY,
      centerX,
      centerY,
      initialElX: el.x,
      initialElY: el.y,
      initialWidth: el.width || 0,
      initialHeight: el.height || 0,
      initialFontSize: el.font_size || 60,
      initialRotation: el.rotation || 0
    };
  };

  const handleDragMove = useCallback((e) => {
    if (!isDragging || selectedElementIdx === null || !currentData) return;

    const res = currentData.resolution || { width: 1920, height: 1080 };
    const container = document.getElementById('preview-img-container');
    if (!container) return;
    const rect = container.getBoundingClientRect();
    const scaleX = res.width / rect.width;
    const scaleY = res.height / rect.height;

    const newData = { ...currentData };
    const el = newData.cards[selectedCardIdx].elements[selectedElementIdx];
    const state = dragRef.current;

    if (state.action === 'move') {
      const dx = (e.clientX - state.startX) * (res.width / rect.width);
      const dy = (e.clientY - state.startY) * (res.height / rect.height);
      el.x = Math.round(state.initialElX + dx);
      el.y = Math.round(state.initialElY + dy);
    } else if (state.action === 'rotate') {
      const dxStart = state.startX - state.centerX;
      const dyStart = state.startY - state.centerY;
      const startAngle = Math.atan2(dyStart, dxStart);

      const dxCurr = e.clientX - state.centerX;
      const dyCurr = e.clientY - state.centerY;
      const currAngle = Math.atan2(dyCurr, dxCurr);

      const angleDiff = (currAngle - startAngle) * (180 / Math.PI);
      el.rotation = Math.round(state.initialRotation + angleDiff);
    } else if (state.action.startsWith('resize-')) {
      const dx = (e.clientX - state.startX) * (res.width / rect.width);
      const dy = (e.clientY - state.startY) * (res.height / rect.height);
      
      // Project dx, dy onto local unrotated axes
      const angle = (state.initialRotation * Math.PI) / 180;
      const localDx = dx * Math.cos(-angle) - dy * Math.sin(-angle);
      const localDy = dx * Math.sin(-angle) + dy * Math.cos(-angle);

      // Determine sign based on which corner is being dragged to achieve center-out scaling
      let signX = 1;
      let signY = 1;
      if (state.action.includes('-tl')) { signX = -1; signY = -1; }
      else if (state.action.includes('-tr')) { signX = 1; signY = -1; }
      else if (state.action.includes('-bl')) { signX = -1; signY = 1; }
      else if (state.action.includes('-br')) { signX = 1; signY = 1; }

      if (el.type === 'text') {
        // For text, just scale font size based on X movement
        el.font_size = Math.max(10, Math.round(state.initialFontSize + (localDx * signX)));
      } else {
        // For rect/circle/image/video, scale width and height from center (hence * 2)
        el.width = Math.max(10, Math.round(state.initialWidth + (localDx * signX * 2)));
        el.height = Math.max(10, Math.round(state.initialHeight + (localDy * signY * 2)));
      }
    }

    setCurrentData(newData);
    
    // Debounce API call heavily during drag to prevent server overload
    clearTimeout(window.dragPreviewTimeout);
    window.dragPreviewTimeout = setTimeout(() => {
      updatePreview(newData);
    }, 100);
  }, [isDragging, selectedElementIdx, currentData]);

  const handleDragEnd = useCallback(() => {
    if (isDragging) {
      setIsDragging(false);
      updatePreview(currentData);
    }
  }, [isDragging, currentData]);

  // Attach global mouse listeners for dragging outside bounds
  useEffect(() => {
    if (isDragging) {
      window.addEventListener('mousemove', handleDragMove);
      window.addEventListener('mouseup', handleDragEnd);
    } else {
      window.removeEventListener('mousemove', handleDragMove);
      window.removeEventListener('mouseup', handleDragEnd);
    }
    return () => {
      window.removeEventListener('mousemove', handleDragMove);
      window.removeEventListener('mouseup', handleDragEnd);
    };
  }, [isDragging, handleDragMove, handleDragEnd]);


  return (
    <div className="app-container" onMouseUp={handleDragEnd}>
      <Sidebar 
        templates={templates} 
        selectedId={selectedId} 
        onSelect={setSelectedId} 
        onNewTemplate={handleNewTemplate}
      />
      
      <div className="main-area">
        <Canvas 
          previewUrl={previewUrl} 
          loading={loading} 
          resolution={currentData?.resolution}
          elements={currentData?.cards?.[selectedCardIdx]?.elements || []}
          selectedElementIdx={selectedElementIdx}
          onSelectElement={setSelectedElementIdx}
          onActionStart={handleActionStart}
        />
        <Timeline 
          cards={currentData?.cards}
          selectedCardIdx={selectedCardIdx}
          onSelectCard={(idx) => {
            setSelectedCardIdx(idx);
            setSelectedElementIdx(null);
            // Re-render preview for the newly selected card. We can simulate it by sending a fake structure where cards[0] is this card.
            // Wait, preview endpoint expects the whole template and always renders card[0].
            // To fix this, we should send the data with the selected card moved to index 0 so the preview endpoint renders it!
            const previewData = { ...currentData, cards: [currentData.cards[idx]] };
            updatePreview(previewData);
          }}
          onAddCard={handleAddCard}
          onRemoveCard={handleRemoveCard}
          onUpdateDuration={handleUpdateCardDuration}
        />
      </div>

      {currentData && (
        <Inspector 
          data={currentData} 
          selectedCardIdx={selectedCardIdx}
          selectedElementIdx={selectedElementIdx}
          onUpdateElement={handleUpdateElement}
          onUpdateTemplate={handleUpdateTemplate}
          onSave={handleSave}
          onPreRender={() => setIsPreRenderOpen(true)}
          onAddElement={handleAddElement}
          onRemoveElement={handleRemoveElement}
          onMoveUp={handleMoveUp}
          onMoveDown={handleMoveDown}
        />
      )}
      
      {isPreRenderOpen && (
        <PreRenderModal 
          data={currentData} 
          onClose={() => setIsPreRenderOpen(false)} 
          onRender={handleFinalRender} 
        />
      )}
    </div>
  )
}

export default App
