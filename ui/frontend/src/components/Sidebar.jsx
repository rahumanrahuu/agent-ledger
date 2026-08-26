import './Sidebar.css'

function Sidebar({ currentView, onViewChange, onSearchClick }) {
  const navItems = [
    { id: 'overview', label: 'Overview', icon: '◆' },
    { id: 'sessions', label: 'Sessions', icon: '●' },
    { id: 'decisions', label: 'Decisions', icon: '◈' },
    { id: 'discoveries', label: 'Discoveries', icon: '▲' },
    { id: 'checkpoints', label: 'Checkpoints', icon: '■' },
    { id: 'timeline', label: 'Timeline', icon: '◇' },
  ]

  return (
    <aside className="sidebar">
      <div className="sidebar-header">
        <div className="app-icon">◦</div>
        <h1>Agent Ledger</h1>
      </div>

      <button className="search-button" onClick={onSearchClick} title="Search (Cmd+K)">
        <span className="search-icon">⌕</span>
        <span className="search-label">Search</span>
        <kbd className="search-shortcut">⌘K</kbd>
      </button>

      <nav className="sidebar-nav">
        {navItems.map((item) => (
          <button
            key={item.id}
            className={`nav-item ${currentView === item.id ? 'active' : ''}`}
            onClick={() => onViewChange(item.id)}
            title={item.label}
          >
            <span className="nav-icon">{item.icon}</span>
            <span className="nav-label">{item.label}</span>
          </button>
        ))}
      </nav>

      <div className="sidebar-footer">
        <div className="status-indicator">
          <span className="status-dot"></span>
          <span className="status-text">Ready</span>
        </div>
      </div>
    </aside>
  )
}

export default Sidebar
