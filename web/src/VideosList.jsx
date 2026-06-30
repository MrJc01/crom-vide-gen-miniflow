import { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';

const API_BASE = `http://${window.location.hostname}:8080/api`;

export default function VideosList() {
  const [jobs, setJobs] = useState({});
  const navigate = useNavigate();

  const loadJobs = () => {
    fetch(`${API_BASE}/videos`)
      .then(r => r.json())
      .then(data => setJobs(data || {}))
      .catch(console.error);
  };

  useEffect(() => {
    loadJobs();
    const interval = setInterval(loadJobs, 3000); // Poll for rendering status
    return () => clearInterval(interval);
  }, []);

  const handleDelete = async (e, id) => {
    e.stopPropagation();
    if (!confirm('Tem certeza que deseja apagar este vídeo permanentemente?')) return;
    
    await fetch(`${API_BASE}/videos/${id}`, { method: 'DELETE' });
    loadJobs();
  };

  const handleArchive = async (e, id, currentArchiveStatus) => {
    e.stopPropagation();
    await fetch(`${API_BASE}/videos/${id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ category: jobs[id].category, archived: !currentArchiveStatus })
    });
    loadJobs();
  };

  const jobList = Object.values(jobs).sort((a, b) => new Date(b.created_at) - new Date(a.created_at));

  return (
    <div style={{ padding: '2rem', overflowY: 'auto' }}>
      <h1 style={{ marginBottom: '2rem' }}>Meus Vídeos Renderizados</h1>

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(300px, 1fr))', gap: '1.5rem' }}>
        {jobList.map(job => (
          <div 
            key={job.id} 
            onClick={() => navigate(`/videos/${job.id}`)}
            style={{ 
              background: 'var(--bg-panel)', padding: '1.5rem', borderRadius: '12px', 
              cursor: 'pointer', border: '1px solid var(--border)',
              opacity: job.archived ? 0.6 : 1
            }}
          >
            <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '1rem' }}>
              <span style={{ 
                background: job.status === 'done' ? 'rgba(34, 197, 94, 0.2)' : (job.status === 'error' ? 'rgba(239, 68, 68, 0.2)' : 'rgba(56, 189, 248, 0.2)'), 
                color: job.status === 'done' ? '#4ade80' : (job.status === 'error' ? '#f87171' : '#38bdf8'),
                padding: '0.2rem 0.5rem', borderRadius: '4px', fontSize: '0.8rem', fontWeight: 'bold'
              }}>
                {job.status === 'done' ? '✅ CONCLUÍDO' : (job.status === 'error' ? '❌ ERRO' : '⏳ RENDERIZANDO...')}
              </span>
              <span style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>
                {new Date(job.created_at).toLocaleString()}
              </span>
            </div>
            
            <h3 style={{ margin: 0, color: '#fff', fontSize: '1.1rem' }}>Template: {job.template_id}</h3>
            <p style={{ color: 'var(--text-muted)', fontSize: '0.9rem', margin: '0.5rem 0 1.5rem 0' }}>
              Job ID: {job.id}
            </p>

            <div style={{ display: 'flex', gap: '0.5rem' }}>
              <button className="btn" style={{ background: '#334155', padding: '0.4rem', fontSize: '0.8rem', width: 'auto' }} onClick={(e) => handleArchive(e, job.id, job.archived)}>
                {job.archived ? '📤 Desarquivar' : '📥 Arquivar'}
              </button>
              <button className="btn" style={{ background: 'transparent', border: '1px solid #ef4444', color: '#ef4444', padding: '0.4rem', fontSize: '0.8rem', width: 'auto' }} onClick={(e) => handleDelete(e, job.id)}>
                🗑️ Excluir
              </button>
            </div>
          </div>
        ))}
        {jobList.length === 0 && <p style={{ color: 'var(--text-muted)' }}>Nenhum vídeo renderizado ainda.</p>}
      </div>
    </div>
  );
}
