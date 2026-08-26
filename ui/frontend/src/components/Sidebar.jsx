import { useState, useEffect } from 'react'
import './Sidebar.css'

function Sidebar({ currentView, onViewChange }) {
  const [repoStatus, setRepoStatus] = useState(null)

  useEffect(() => {
    const fetchRepoStatus = async () => {
      try {
        const response = await fetch('/api/overview')
        if (response.ok) {
          const data = await response.json()
          setRepoStatus(data)
        }
      } catch (err) {
        console.error('Failed to fetch repo status:', err)
      }
    }

    fetchRepoStatus()
  }, [])

  const navItems = [
    { section: 'AGENT', items: [{ id: 'overview', label: 'Overview' }] },
    {
      section: 'HISTORY',
      items: [
        { id: 'sessions', label: 'Sessions' },
        { id: 'timeline', label: 'Timeline' },
      ],
    },
    {
      section: 'KNOWLEDGE',
      items: [
        { id: 'decisions', label: 'Decisions' },
        { id: 'discoveries', label: 'Discoveries' },
        { id: 'checkpoints', label: 'Checkpoints' },
        { id: 'knowledge-graph', label: 'Knowledge Graph' },
      ],
    },
    {
      section: 'PROJECT',
      items: [
        { id: 'files', label: 'Files' },
        { id: 'metadata', label: 'Metadata' },
      ],
    },
  ]

  return (
    <aside className="sidebar">
      <div className="sidebar-header">
        <div className="app-title">
          <span className="app-icon">◦</span>
          <span className="app-name">Agent Ledger</span>
        </div>
        {repoStatus && (
          <div className="repo-branch">{repoStatus.current_branch}</div>
        )}
      </div>

      <nav className="sidebar-nav">
        {navItems.map((section) => (
          <div key={section.section} className="nav-section">
            <h3 className="section-title">{section.section}</h3>
            <ul className="section-items">
              {section.items.map((item) => (
                <li key={item.id}>
                  <button
                    className={`nav-item ${
                      currentView === item.id ? 'active' : ''
                    }`}
                    onClick={() => onViewChange(item.id)}
                  >
                    {item.label}
                  </button>
                </li>
              ))}
            </ul>
          </div>
        ))}
      </nav>

      <div className="sidebar-footer">
        {repoStatus && (
          <div className="repo-info">
            <div className="repo-name">{repoStatus.project_name}</div>
            <div className="repo-details">
              {repoStatus.current_branch} • {repoStatus.current_commit?.substring(0, 7)}
            </div>
          </div>
        )}
        <div className="status-indicator">
          <span className="status-dot"></span>
          <span className="status-label">Ready</span>
        </div>
      </div>
    </aside>
  )
}

export default Sidebar
