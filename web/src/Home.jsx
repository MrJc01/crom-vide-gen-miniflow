import { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';

const API_BASE = 'http://localhost:8080/api';

export default function Home() {
  const [templates, setTemplates] = useState([]);
  const [search, setSearch] = useState('');
  const navigate = useNavigate();

  useEffect(() => {
    fetch(`${API_BASE}/templates`)
      .then(r => r.json())
      .then(data => setTemplates(data || []))
      .catch(console.error);
  }, []);

  const filtered = templates.filter(t => t.toLowerCase().includes(search.toLowerCase()));

  return (
    <div style={{ padding: '2rem', overflowY: 'auto' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '2rem' }}>
        <h1>Templates</h1>
        <input 
          type="text" 
          placeholder="Pesquisar..." 
          value={search}
          onChange={e => setSearch(e.target.value)}
          style={{ padding: '0.8rem', width: '300px', borderRadius: '8px', border: '1px solid var(--border)', background: 'var(--bg-panel)', color: '#fff' }}
        />
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(250px, 1fr))', gap: '1.5rem' }}>
        {filtered.map(t => (
          <div key={t} onClick={() => navigate(`/editor/${t}`)} style={{ background: 'var(--bg-panel)', padding: '1.5rem', borderRadius: '12px', cursor: 'pointer', border: '1px solid var(--border)', transition: 'transform 0.2s, borderColor 0.2s' }} className="template-card">
            <div style={{ fontSize: '2rem', marginBottom: '1rem' }}>📄</div>
            <h3 style={{ margin: 0, color: 'var(--accent)' }}>{t}</h3>
            <p style={{ color: 'var(--text-muted)', fontSize: '0.9rem', marginTop: '0.5rem' }}>Clique para editar o design e gerar um vídeo.</p>
          </div>
        ))}
        {filtered.length === 0 && <p style={{ color: 'var(--text-muted)' }}>Nenhum template encontrado.</p>}
      </div>
    </div>
  );
}
