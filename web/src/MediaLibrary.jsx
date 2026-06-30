import { useState, useEffect } from 'react';

const API_HOST = `http://${window.location.hostname}:8080`;

export default function MediaLibrary({ onSelect, isPicker = false, onClose }) {
  const [mediaTree, setMediaTree] = useState({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [filterType, setFilterType] = useState('all'); // 'all' | 'video' | 'image' | 'audio'
  const [selectedFolder, setSelectedFolder] = useState('/');
  const [uploading, setUploading] = useState(false);

  const fetchMedia = async () => {
    try {
      setLoading(true);
      const res = await fetch(`${API_HOST}/api/media`);
      if (!res.ok) throw new Error('Erro ao carregar galeria de mídias');
      const data = await res.json();
      setMediaTree(data);
      
      // Se a pasta anteriormente selecionada não existir mais, volta pra raiz
      if (data && !data[selectedFolder] && Object.keys(data).length > 0) {
        setSelectedFolder(Object.keys(data)[0]);
      }
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchMedia();
  }, []);

  const handleFileUpload = async (e) => {
    const file = e.target.files[0];
    if (!file) return;

    const formData = new FormData();
    formData.append('file', file);
    
    // Opcional: Se quisermos suportar envio para a pasta atual selecionada
    // Mas por padrão o /api/upload salva na raiz de uploads.

    try {
      setUploading(true);
      const res = await fetch(`${API_HOST}/api/upload`, {
        method: 'POST',
        body: formData
      });
      if (!res.ok) throw new Error('Falha ao subir arquivo');
      await fetchMedia();
      alert('Mídia enviada com sucesso!');
    } catch (err) {
      alert('Erro no upload: ' + err.message);
    } finally {
      setUploading(false);
    }
  };

  const copyToClipboard = (text) => {
    navigator.clipboard.writeText(text);
    alert('Caminho copiado para a área de transferência!');
  };

  // Processa itens da pasta ativa filtrando por busca e tipo
  const getFilteredItems = () => {
    const folderItems = mediaTree[selectedFolder] || [];
    return folderItems.filter(item => {
      const matchesSearch = item.name.toLowerCase().includes(searchQuery.toLowerCase());
      const matchesType = filterType === 'all' || item.type === filterType;
      return matchesSearch && matchesType;
    });
  };

  const folders = Object.keys(mediaTree).sort();
  const currentItems = getFilteredItems();

  return (
    <div className="media-library-container" style={containerStyle(isPicker)}>
      {isPicker && (
        <div style={pickerHeaderStyle}>
          <h3>Biblioteca de Mídias</h3>
          <button onClick={onClose} style={closeBtnStyle}>×</button>
        </div>
      )}

      {!isPicker && (
        <div style={pageHeaderStyle}>
          <div>
            <h1 style={{ margin: 0, fontSize: '1.8rem', color: '#f8fafc' }}>Biblioteca de Mídias</h1>
            <p style={{ margin: '0.2rem 0 0 0', color: 'var(--text-muted)', fontSize: '0.9rem' }}>
              Gerencie e selecione todos os assets de vídeos, imagens e trilhas sonoras.
            </p>
          </div>
          
          <label className="btn" style={uploadBtnStyle(uploading)}>
            {uploading ? '⬆️ Enviando...' : '➕ Enviar Nova Mídia'}
            <input type="file" style={{ display: 'none' }} onChange={handleFileUpload} disabled={uploading} />
          </label>
        </div>
      )}

      <div style={mainLayoutStyle}>
        {/* Sidebar - Pastas */}
        <div style={sidebarStyle}>
          <h4 style={{ margin: '0 0 0.8rem 0', color: '#94a3b8', fontSize: '0.8rem', textTransform: 'uppercase' }}>Pastas</h4>
          {loading && <div style={{ fontSize: '0.85rem', color: 'var(--text-muted)' }}>Carregando...</div>}
          {!loading && folders.length === 0 && (
            <div style={{ fontSize: '0.85rem', color: 'var(--text-muted)', fontStyle: 'italic' }}>Nenhuma pasta.</div>
          )}
          {folders.map(folder => (
            <button 
              key={folder} 
              onClick={() => setSelectedFolder(folder)}
              style={folderBtnStyle(selectedFolder === folder)}
            >
              📁 {folder === '/' ? 'raiz' : folder}
              <span style={folderBadgeStyle(selectedFolder === folder)}>
                {mediaTree[folder]?.length || 0}
              </span>
            </button>
          ))}
        </div>

        {/* Content Area - Mídias */}
        <div style={contentAreaStyle}>
          {/* Top Bar - Filtros e Busca */}
          <div style={filterBarStyle}>
            <input 
              type="text" 
              placeholder="Pesquisar por nome..." 
              value={searchQuery}
              onChange={e => setSearchQuery(e.target.value)}
              style={searchInputStyle}
            />

            <div style={{ display: 'flex', gap: '0.3rem' }}>
              {['all', 'video', 'image', 'audio'].map(type => (
                <button 
                  key={type} 
                  onClick={() => setFilterType(type)}
                  style={filterBtnStyle(filterType === type)}
                >
                  {type === 'all' ? 'Todos' : (type === 'video' ? '🎥 Vídeos' : (type === 'image' ? '🖼️ Imagens' : '🎵 Áudio'))}
                </button>
              ))}
            </div>

            {isPicker && (
              <label className="btn" style={{ ...uploadBtnStyle(uploading), fontSize: '0.8rem', padding: '0.4rem 0.8rem' }}>
                {uploading ? 'Enviando...' : '➕ Enviar'}
                <input type="file" style={{ display: 'none' }} onChange={handleFileUpload} disabled={uploading} />
              </label>
            )}
          </div>

          {/* Grid de Itens */}
          {loading ? (
            <div style={centerMessageStyle}>Buscando arquivos no disco...</div>
          ) : error ? (
            <div style={{ ...centerMessageStyle, color: '#ef4444' }}>{error}</div>
          ) : currentItems.length === 0 ? (
            <div style={centerMessageStyle}>Nenhum arquivo encontrado nesta pasta com estes filtros.</div>
          ) : (
            <div style={gridStyle}>
              {currentItems.map(item => (
                <div key={item.path} style={cardStyle} onClick={() => onSelect && onSelect(item.path)}>
                  {/* Preview Container */}
                  <div style={previewContainerStyle}>
                    {item.type === 'image' && (
                      <img src={item.url} alt={item.name} style={imgPreviewStyle} />
                    )}
                    {item.type === 'video' && (
                      <video src={item.url} style={videoPreviewStyle} preload="metadata" muted />
                    )}
                    {item.type === 'audio' && (
                      <div style={audioIconStyle}>🎵</div>
                    )}
                    {item.type === 'other' && (
                      <div style={audioIconStyle}>🧩</div>
                    )}

                    {/* Badge de Tipo */}
                    <span style={typeBadgeStyle(item.type)}>{item.type}</span>
                  </div>

                  {/* Detalhes do Card */}
                  <div style={cardFooterStyle}>
                    <div style={nameStyle} title={item.name}>{item.name}</div>
                    <div style={sizeStyle}>{(item.size / (1024 * 1024)).toFixed(2)} MB</div>
                    
                    {/* Ações */}
                    <div style={{ display: 'flex', gap: '0.4rem', marginTop: '0.4rem' }}>
                      {onSelect ? (
                        <button style={selectActionBtnStyle}>Selecionar</button>
                      ) : (
                        <>
                          <button onClick={(e) => { e.stopPropagation(); copyToClipboard(item.path); }} style={actionBtnStyle}>
                            Copiar Caminho Local
                          </button>
                          <a href={item.url} target="_blank" rel="noreferrer" onClick={e => e.stopPropagation()} style={actionLinkStyle}>
                            Abrir URL
                          </a>
                        </>
                      )}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

// Inline Styles
const containerStyle = (isPicker) => ({
  display: 'flex',
  flexDirection: 'column',
  width: '100%',
  height: '100%',
  background: isPicker ? 'var(--bg-panel)' : 'transparent',
  color: 'var(--text-main)',
  fontFamily: 'inherit',
  overflow: 'hidden'
});

const pickerHeaderStyle = {
  padding: '1rem',
  borderBottom: '1px solid var(--border)',
  display: 'flex',
  justifyContent: 'space-between',
  alignItems: 'center',
  background: 'rgba(0,0,0,0.1)'
};

const closeBtnStyle = {
  background: 'none',
  border: 'none',
  color: 'var(--text-muted)',
  fontSize: '1.5rem',
  cursor: 'pointer'
};

const pageHeaderStyle = {
  display: 'flex',
  justifyContent: 'space-between',
  alignItems: 'center',
  marginBottom: '1.5rem',
  paddingBottom: '1rem',
  borderBottom: '1px solid var(--border)'
};

const uploadBtnStyle = (uploading) => ({
  cursor: uploading ? 'not-allowed' : 'pointer',
  background: 'var(--accent)',
  color: '#000',
  fontWeight: 'bold',
  padding: '0.6rem 1.2rem',
  borderRadius: '6px',
  width: 'auto',
  display: 'inline-flex',
  alignItems: 'center',
  fontSize: '0.9rem',
  opacity: uploading ? 0.7 : 1
});

const mainLayoutStyle = {
  display: 'flex',
  flex: 1,
  overflow: 'hidden',
  height: '100%',
  minHeight: '400px'
};

const sidebarStyle = {
  width: '200px',
  borderRight: '1px solid var(--border)',
  padding: '1rem',
  overflowY: 'auto',
  background: 'rgba(0,0,0,0.05)',
  display: 'flex',
  flexDirection: 'column',
  gap: '0.3rem'
};

const folderBtnStyle = (isActive) => ({
  width: '100%',
  padding: '0.6rem 0.8rem',
  borderRadius: '6px',
  background: isActive ? 'rgba(56, 189, 248, 0.15)' : 'transparent',
  color: isActive ? 'var(--accent)' : 'var(--text-main)',
  border: isActive ? '1px solid var(--accent)' : '1px solid transparent',
  textAlign: 'left',
  fontSize: '0.85rem',
  cursor: 'pointer',
  display: 'flex',
  justifyContent: 'space-between',
  alignItems: 'center',
  transition: 'all 0.2s'
});

const folderBadgeStyle = (isActive) => ({
  fontSize: '0.75rem',
  padding: '0.1rem 0.4rem',
  borderRadius: '99px',
  background: isActive ? 'var(--accent)' : 'var(--border)',
  color: isActive ? '#000' : 'var(--text-muted)',
  fontWeight: 'bold'
});

const contentAreaStyle = {
  flex: 1,
  padding: '1rem',
  display: 'flex',
  flexDirection: 'column',
  overflowY: 'auto'
};

const filterBarStyle = {
  display: 'flex',
  gap: '1rem',
  alignItems: 'center',
  marginBottom: '1rem',
  flexWrap: 'wrap'
};

const searchInputStyle = {
  flex: 1,
  minWidth: '200px',
  padding: '0.5rem 0.75rem',
  borderRadius: '6px',
  background: 'var(--bg-main)',
  border: '1px solid var(--border)',
  color: 'var(--text-main)',
  fontSize: '0.85rem'
};

const filterBtnStyle = (isActive) => ({
  padding: '0.4rem 0.8rem',
  borderRadius: '6px',
  background: isActive ? '#334155' : 'transparent',
  color: isActive ? '#fff' : 'var(--text-muted)',
  border: isActive ? '1px solid var(--border)' : '1px solid transparent',
  fontSize: '0.8rem',
  cursor: 'pointer',
  transition: 'all 0.2s'
});

const centerMessageStyle = {
  display: 'flex',
  justifyContent: 'center',
  alignItems: 'center',
  flex: 1,
  color: 'var(--text-muted)',
  fontStyle: 'italic',
  fontSize: '0.9rem',
  padding: '3rem 0'
};

const gridStyle = {
  display: 'grid',
  gridTemplateColumns: 'repeat(auto-fill, minmax(160px, 1fr))',
  gap: '1rem',
  overflowY: 'auto'
};

const cardStyle = {
  background: 'var(--bg-main)',
  border: '1px solid var(--border)',
  borderRadius: '8px',
  overflow: 'hidden',
  cursor: 'pointer',
  transition: 'transform 0.2s, border-color 0.2s',
  display: 'flex',
  flexDirection: 'column',
  position: 'relative'
};

const previewContainerStyle = {
  height: '100px',
  background: '#090d16',
  display: 'flex',
  justifyContent: 'center',
  alignItems: 'center',
  overflow: 'hidden',
  position: 'relative'
};

const imgPreviewStyle = {
  width: '100%',
  height: '100%',
  objectFit: 'cover'
};

const videoPreviewStyle = {
  width: '100%',
  height: '100%',
  objectFit: 'cover'
};

const audioIconStyle = {
  fontSize: '2.5rem',
  opacity: 0.7
};

const typeBadgeStyle = (type) => ({
  position: 'absolute',
  top: '5px',
  right: '5px',
  fontSize: '0.65rem',
  textTransform: 'uppercase',
  padding: '0.1rem 0.3rem',
  borderRadius: '3px',
  fontWeight: 'bold',
  background: type === 'video' ? '#ef4444' : (type === 'image' ? '#3b82f6' : '#10b981'),
  color: '#fff'
});

const cardFooterStyle = {
  padding: '0.6rem',
  display: 'flex',
  flexDirection: 'column',
  flex: 1,
  justifyContent: 'space-between'
};

const nameStyle = {
  fontSize: '0.8rem',
  fontWeight: '600',
  whiteSpace: 'nowrap',
  overflow: 'hidden',
  textOverflow: 'ellipsis',
  color: '#e2e8f0'
};

const sizeStyle = {
  fontSize: '0.7rem',
  color: 'var(--text-muted)',
  marginTop: '0.1rem'
};

const selectActionBtnStyle = {
  width: '100%',
  padding: '0.4rem',
  fontSize: '0.75rem',
  background: 'var(--accent)',
  color: '#000',
  border: 'none',
  borderRadius: '4px',
  cursor: 'pointer',
  fontWeight: 'bold'
};

const actionBtnStyle = {
  flex: 1,
  padding: '0.3rem 0.5rem',
  fontSize: '0.7rem',
  background: '#334155',
  color: '#f1f5f9',
  border: 'none',
  borderRadius: '4px',
  cursor: 'pointer',
  transition: 'background 0.2s'
};

const actionLinkStyle = {
  padding: '0.3rem 0.5rem',
  fontSize: '0.7rem',
  background: 'transparent',
  color: 'var(--accent)',
  border: '1px solid var(--accent)',
  borderRadius: '4px',
  textAlign: 'center',
  textDecoration: 'none',
  cursor: 'pointer',
  flex: 1
};
