import './Inspector.css'

function Inspector({ item, onClose }) {
  if (!item) return null

  return (
    <aside className="inspector">
      <div className="inspector-header">
        <h2>Inspector</h2>
        <button className="inspector-close" onClick={onClose}>✕</button>
      </div>

      <div className="inspector-content">
        <div className="inspector-section">
          <h3 className="section-header">Type</h3>
          <div className="inspector-value">{item.type}</div>
        </div>

        {item.title && (
          <div className="inspector-section">
            <h3 className="section-header">Title</h3>
            <div className="inspector-value">{item.title}</div>
          </div>
        )}

        {item.id && (
          <div className="inspector-section">
            <h3 className="section-header">ID</h3>
            <code className="inspector-code">{item.id}</code>
          </div>
        )}

        {item.timestamp && (
          <div className="inspector-section">
            <h3 className="section-header">Timestamp</h3>
            <div className="inspector-value">
              {new Date(item.timestamp).toLocaleString()}
            </div>
          </div>
        )}

        {item.path && (
          <div className="inspector-section">
            <h3 className="section-header">Path</h3>
            <code className="inspector-code">{item.path}</code>
          </div>
        )}

        {item.content && (
          <div className="inspector-section">
            <h3 className="section-header">Content</h3>
            <div className="inspector-markdown">
              {item.content.substring(0, 300)}...
            </div>
          </div>
        )}
        {item.value !== undefined && (
          <div className="inspector-section"><h3 className="section-header">Value</h3><div className="inspector-value">{String(item.value)}</div></div>
        )}
        {item.connected?.length > 0 && (
          <div className="inspector-section"><h3 className="section-header">Connected</h3><div className="relationship-list">
            {item.connected.map((relationship, index) => <div className="relationship" key={`${relationship.source}-${relationship.target}-${index}`}><span className={`relationship-${relationship.node?.type || 'other'}`} /><div><strong>{relationship.node?.label || relationship.node?.id}</strong><small>{relationship.type || 'related to'} · {relationship.node?.type}</small></div></div>)}
          </div></div>
        )}
      </div>
    </aside>
  )
}

export default Inspector
