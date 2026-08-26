import './CommonView.css'

function Discoveries({ onSelect }) {
  return (
    <div className="view-container">
      <header className="view-header">
        <h1>Discoveries</h1>
        <p className="text-secondary">Review all recorded discoveries</p>
      </header>
      <div className="empty-state">
        <p>Discoveries view coming soon</p>
      </div>
    </div>
  )
}

export default Discoveries
