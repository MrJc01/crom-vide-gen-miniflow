import { useState, useEffect } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';

const HOST = `http://${window.location.hostname}:8080`;
const API_BASE = `${HOST}/api`;

export default function VideoDetail() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [job, setJob] = useState(null);
  const [error, setError] = useState(null);
  const [showJson, setShowJson] = useState(false);

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
            
            <div style={{ marginTop: '2rem', width: '100%', maxWidth: '800px', display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
              <div style={{ color: 'var(--text-muted)', fontSize: '0.9rem' }}>
                <p><strong>Criado em:</strong> {new Date(job.created_at).toLocaleString()}</p>
                {job.render_duration_ms > 0 && (
                  <p><strong>Tempo de Renderização:</strong> {(job.render_duration_ms / 1000).toFixed(2)}s</p>
                )}
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

            {job.template && (
              <div style={{ marginTop: '1.5rem', width: '100%', maxWidth: '800px', borderTop: '1px solid var(--border)', paddingTop: '1.5rem' }}>
                <button 
                  onClick={() => setShowJson(!showJson)}
                  style={{
                    background: 'none', border: 'none', color: 'var(--accent)', cursor: 'pointer', fontSize: '0.95rem', fontWeight: 'bold', display: 'flex', alignItems: 'center', gap: '0.5rem', padding: 0
                  }}
                >
                  {showJson ? '▼ Ocultar JSON de Template Utilizado' : '▶ Exibir JSON de Template Utilizado'}
                </button>
                {showJson && (
                  <div style={{ position: 'relative', marginTop: '1rem' }}>
                    <button 
                      onClick={() => {
                        navigator.clipboard.writeText(JSON.stringify(job.template, null, 2));
                        alert('JSON copiado!');
                      }}
                      style={{
                        position: 'absolute', top: '10px', right: '10px', background: '#475569', color: '#fff', border: 'none', borderRadius: '4px', padding: '0.3rem 0.6rem', fontSize: '0.75rem', cursor: 'pointer'
                      }}
                    >
                      Copiar JSON
                    </button>
                    <pre style={{
                      background: 'var(--bg-main)', border: '1px solid var(--border)', borderRadius: '6px', padding: '1rem', overflowX: 'auto', fontSize: '0.8rem', color: '#cbd5e1', maxHeight: '300px', margin: 0
                    }}>
                      <code>{JSON.stringify(job.template, null, 2)}</code>
                    </pre>
                  </div>
                )}
              </div>
            )}
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
