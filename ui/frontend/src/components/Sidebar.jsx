import { useState, useEffect } from 'react'
import {
  FiBookOpen,
  FiSearch,
  FiGrid,
  FiTerminal,
  FiCheckSquare,
  FiCompass,
  FiBookmark,
  FiClock,
} from 'react-icons/fi'
import './Sidebar.css'

function Sidebar({ currentView, onViewChange, onSearchClick }) {
  const [isMac, setIsMac] = useState(false)

  useEffect(() => {
    setIsMac(
      typeof navigator !== 'undefined' &&
        (navigator.platform?.toUpperCase().indexOf('MAC') >= 0 ||
          navigator.userAgent?.includes('Mac'))
    )
  }, [])

  const navItems = [
    { id: 'overview', label: 'Overview', icon: <FiGrid /> },
    { id: 'sessions', label: 'Sessions', icon: <FiTerminal /> },
    { id: 'decisions', label: 'Decisions', icon: <FiCheckSquare /> },
    { id: 'discoveries', label: 'Discoveries', icon: <FiCompass /> },
    { id: 'checkpoints', label: 'Checkpoints', icon: <FiBookmark /> },
    { id: 'timeline', label: 'Timeline', icon: <FiClock /> },
  ]

  return (
    <aside className="sidebar">
      <div className="sidebar-header">
        <div className="app-icon">
          <FiBookOpen />
        </div>
        <h1>Agent Ledger</h1>
      </div>

      <button
        className="search-button"
        onClick={onSearchClick}
        title={`Search (${isMac ? 'Cmd+K' : 'Ctrl+K'})`}
      >
        <span className="search-icon">
          <FiSearch />
        </span>
        <span className="search-label">Search</span>
        <kbd className="search-shortcut">{isMac ? '⌘K' : 'Ctrl+K'}</kbd>
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
