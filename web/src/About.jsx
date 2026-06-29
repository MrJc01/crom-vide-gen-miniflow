export default function About() {
  return (
    <div style={{ padding: '2rem', maxWidth: '800px', margin: '0 auto', textAlign: 'center' }}>
      <h1 style={{ color: 'var(--accent)', marginBottom: '1rem', fontSize: '3rem' }}>VideGen Plataforma</h1>
      <p style={{ fontSize: '1.2rem', color: 'var(--text-muted)', marginBottom: '3rem' }}>
        A evolução da geração de vídeos programática.
      </p>

      <div style={{ background: 'var(--bg-panel)', padding: '2rem', borderRadius: '12px', border: '1px solid var(--border)' }}>
        <h3 style={{ marginBottom: '1rem' }}>Tecnologias Utilizadas</h3>
        <ul style={{ listStyle: 'none', padding: 0, display: 'flex', flexDirection: 'column', gap: '0.8rem', color: 'var(--text-muted)' }}>
          <li>🟢 Go (Golang) - Backend ultra rápido e orquestração do FFmpeg</li>
          <li>⚛️ React + Vite - Frontend dinâmico estilo SPA e gerenciamento visual</li>
          <li>🎞️ FFmpeg / gg - Motor visual e unificação de mídias por baixo dos panos</li>
        </ul>
      </div>
    </div>
  );
}
