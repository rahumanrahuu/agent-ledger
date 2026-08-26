import './Inspector.css'

function Inspector({ data, onClose }) {
  if (!data) return null

  return (
    <aside className="inspector">
      <div className="inspector-header">
        <h2>Details</h2>
        <button className="close-button" onClick={onClose} title="Close Inspector">
          ✕
        </button>
      </div>

      <div className="inspector-content">
        {data.type === 'session' && (
          <div className="inspector-section">
            <h3>{data.agent || 'Session'}</h3>
            <div className="property-list">
              <div className="property">
                <span className="property-label">ID</span>
                <span className="property-value font-mono">{data.id}</span>
              </div>
              <div className="property">
                <span className="property-label">Branch</span>
                <span className="property-value">{data.branch}</span>
              </div>
              <div className="property">
                <span className="property-label">Head</span>
                <span className="property-value font-mono">{data.head?.substring(0, 7)}</span>
              </div>
              <div className="property">
                <span className="property-label">Started</span>
                <span className="property-value">
                  {new Date(data.startTime).toLocaleString()}
                </span>
              </div>
              {data.endTime && (
                <div className="property">
                  <span className="property-label">Ended</span>
                  <span className="property-value">
                    {new Date(data.endTime).toLocaleString()}
                  </span>
                </div>
              )}
            </div>
          </div>
        )}

        {data.type === 'decision' && (
          <div className="inspector-section">
            <h3>{data.title}</h3>
            <div className="property-list">
              <div className="property">
                <span className="property-label">Type</span>
                <span className="property-value">Decision</span>
              </div>
              <div className="property full-width">
                <span className="property-label">Content</span>
                <p className="property-value">{data.content}</p>
              </div>
            </div>
          </div>
        )}

        {!data.type && (
          <div className="inspector-section">
            <div className="property-list">
              <div className="property full-width">
                <pre>{JSON.stringify(data, null, 2)}</pre>
              </div>
            </div>
          </div>
        )}
      </div>
    </aside>
  )
}

export default Inspector
