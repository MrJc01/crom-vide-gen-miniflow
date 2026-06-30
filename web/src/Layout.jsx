import { Link } from 'react-router-dom';

export default function Layout({ children }) {
  return (
    <div className="layout-container" style={{ display: 'flex', height: '100vh', overflow: 'hidden' }}>
      <aside style={{ width: '250px', background: 'var(--bg-panel)', borderRight: '1px solid var(--border)', display: 'flex', flexDirection: 'column' }}>
        <div style={{ padding: '1.5rem', borderBottom: '1px solid var(--border)', fontWeight: 'bold', fontSize: '1.2rem', color: 'var(--accent)' }}>
          VideGen
        </div>
        <nav style={{ flex: 1, padding: '1rem', display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
          <Link to="/" className="nav-link" style={linkStyle}>🏠 Início</Link>
          <Link to="/videos" className="nav-link" style={linkStyle}>🎥 Meus Vídeos</Link>
          <Link to="/medias" className="nav-link" style={linkStyle}>📂 Mídias</Link>
          <Link to="/docs" className="nav-link" style={linkStyle}>📚 Documentação</Link>
          <Link to="/about" className="nav-link" style={linkStyle}>ℹ️ Sobre</Link>
        </nav>
      </aside>
      <main style={{ flex: 1, display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
        {children}
      </main>
    </div>
  );
}

const linkStyle = {
  padding: '0.8rem 1rem',
  color: 'var(--text-main)',
  textDecoration: 'none',
  borderRadius: '8px',
  display: 'block',
  transition: 'background 0.2s',
  border: '1px solid transparent'
};
