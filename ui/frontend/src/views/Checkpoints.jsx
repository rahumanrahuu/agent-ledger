import './CommonView.css'

function Checkpoints({ onSelect }) {
  return (
    <div className="view-container">
      <header className="view-header">
        <h1>Checkpoints</h1>
        <p className="text-secondary">Review all recorded checkpoints</p>
      </header>
      <div className="empty-state">
        <p>Checkpoints view coming soon</p>
      </div>
    </div>
  )
}

export default Checkpoints
