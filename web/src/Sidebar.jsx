export default function Sidebar({ templates, selectedId, onSelect, onNewTemplate }) {
  const handleNew = () => {
    const name = prompt('Nome do novo template (sem .json):');
    if (name) {
      onNewTemplate(name);
    }
  };

  return (
    <div className="sidebar">
      <div className="panel-header">
        Templates
        <button 
          className="btn" 
          style={{ padding: '0.4rem 0.6rem', fontSize: '0.8rem', width: 'auto' }} 
          onClick={handleNew}
        >
          ✨ Novo
        </button>
      </div>
      <div className="panel-content">
        {templates.map(id => (
          <div 
            key={id} 
            className={`template-item ${selectedId === id ? 'active' : ''}`}
            onClick={() => onSelect(id)}
          >
            📄 {id}.json
          </div>
        ))}
      </div>
    </div>
  )
}
