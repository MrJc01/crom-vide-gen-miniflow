import { useState, useEffect } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';

const API_BASE = 'http://localhost:8080/api';
const HOST = 'http://localhost:8080';

export default function VideoDetail() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [job, setJob] = useState(null);
  const [error, setError] = useState(null);

  useEffect(() => {
    let interval;
    const fetchJob = () => {
      fetch(`${API_BASE}/videos/${id}`)
        .then(r => {
          if (!r.ok) throw new Error('Vídeo não encontrado');
          return r.json();
        })
        .then(data => {
          setJob(data);
          if (data.status !== 'rendering' && interval) {
            clearInterval(interval);
          }
        })
        .catch(err => {
          setError(err.message);
          if (interval) clearInterval(interval);
        });
    };

    fetchJob();
    interval = setInterval(fetchJob, 2000);
    return () => clearInterval(interval);
  }, [id]);

  if (error) return <div style={{ padding: '2rem', color: '#f87171' }}>Erro: {error}</div>;
  if (!job) return <div style={{ padding: '2rem' }}>Carregando dados do vídeo...</div>;

  return (
    <div style={{ padding: '2rem', maxWidth: '1000px', margin: '0 auto', overflowY: 'auto', width: '100%' }}>
      <Link to="/videos" style={{ color: 'var(--accent)', textDecoration: 'none', marginBottom: '1rem', display: 'inline-block' }}>
        ← Voltar para Meus Vídeos
      </Link>
      
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '2rem' }}>
        <h1 style={{ margin: 0 }}>Vídeo: {job.template_id}</h1>
        <button 
          className="btn" 
          style={{ background: 'var(--accent)', color: '#000', width: 'auto' }}
          onClick={() => navigate(`/editor/${job.template_id}`)}
        >
          ✏️ Editar Template e Gerar Novo
        </button>
      </div>

      <div style={{ background: 'var(--bg-panel)', padding: '2rem', borderRadius: '12px', border: '1px solid var(--border)' }}>
        
        {job.status === 'rendering' && (
          <div style={{ textAlign: 'center', padding: '4rem 0' }}>
            <div style={{ fontSize: '3rem', marginBottom: '1rem', animation: 'spin 2s linear infinite' }}>⏳</div>
            <h2 style={{ color: 'var(--accent)' }}>Renderizando o Vídeo...</h2>
            <p style={{ color: 'var(--text-muted)' }}>O motor FFmpeg está processando as cenas e injetando as variáveis no backend. Por favor, aguarde.</p>
          </div>
        )}

        {job.status === 'error' && (
          <div style={{ textAlign: 'center', padding: '2rem 0', color: '#f87171' }}>
            <h2>❌ Ocorreu um Erro na Renderização</h2>
            <p>{job.error || 'Erro desconhecido.'}</p>
          </div>
        )}

        {job.status === 'done' && (
          <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
            <video 
              controls 
              style={{ width: '100%', maxWidth: '800px', borderRadius: '8px', border: '1px solid var(--border)', background: '#000' }}
              src={`${HOST}/${job.file_path}`}
            >
              Seu navegador não suporta vídeos HTML5.
            </video>
            
            <div style={{ marginTop: '2rem', width: '100%', maxWidth: '800px', display: 'flex', justifyContent: 'space-between' }}>
              <div style={{ color: 'var(--text-muted)', fontSize: '0.9rem' }}>
                <p><strong>Criado em:</strong> {new Date(job.created_at).toLocaleString()}</p>
                <p><strong>ID do Job:</strong> {job.id}</p>
              </div>
              <div>
                <a 
                  href={`${HOST}/${job.file_path}`} 
                  download 
                  className="btn" 
                  style={{ background: '#334155', color: '#fff', textDecoration: 'none', display: 'inline-block', padding: '0.6rem 1rem' }}
                >
                  ⬇️ Baixar MP4
                </a>
              </div>
            </div>
          </div>
        )}

      </div>
      <style>{`
        @keyframes spin {
          100% { transform: rotate(360deg); }
        }
      `}</style>
    </div>
  );
}
