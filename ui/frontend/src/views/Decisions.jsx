import './CommonView.css'

function Decisions({ onSelect }) {
  return (
    <div className="view-container">
      <header className="view-header">
        <h1>Decisions</h1>
        <p className="text-secondary">Review all recorded decisions</p>
      </header>
      <div className="empty-state">
        <p>Decisions view coming soon</p>
      </div>
    </div>
  )
}

export default Decisions
