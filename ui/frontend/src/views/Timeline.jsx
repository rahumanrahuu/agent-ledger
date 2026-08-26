import './CommonView.css'

function Timeline({ onSelect }) {
  return (
    <div className="view-container">
      <header className="view-header">
        <h1>Timeline</h1>
        <p className="text-secondary">Unified chronological stream of all events</p>
      </header>
      <div className="empty-state">
        <p>Timeline view coming soon</p>
      </div>
    </div>
  )
}

export default Timeline
