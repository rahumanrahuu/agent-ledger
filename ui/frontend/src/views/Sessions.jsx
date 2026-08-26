import { useState, useEffect } from 'react'
import './Sessions.css'

function Sessions({ onSelect }) {
  const [sessions, setSessions] = useState([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const fetchSessions = async () => {
      try {
        const response = await fetch('/api/sessions')
        if (!response.ok) throw new Error('Failed to fetch sessions')
        const data = await response.json()
        setSessions(data.sessions || [])
      } catch (err) {
        console.error(err)
      } finally {
        setLoading(false)
      }
    }

    fetchSessions()
  }, [])

  if (loading) return <div className="view-container">Loading sessions...</div>

  return (
    <div className="view-container">
      <header className="view-header">
        <h1>Sessions</h1>
        <p className="text-secondary">{sessions.length} sessions found</p>
      </header>

      <div className="sessions-list">
        {sessions.length === 0 ? (
          <div className="empty-state">
            <p>No sessions yet</p>
          </div>
        ) : (
          sessions.map((session) => (
            <button
              key={session.id}
              className="session-card"
              onClick={() => onSelect({ type: 'session', ...session })}
            >
              <div className="session-header">
                <div className="session-name">{session.agent || 'Session'}</div>
                <div className={`session-status ${session.status}`}>
                  {session.status}
                </div>
              </div>
              <div className="session-meta">
                <span className="meta-item">
                  <span className="meta-label">Branch:</span>
                  <span className="meta-value font-mono">{session.branch}</span>
                </span>
                <span className="meta-item">
                  <span className="meta-label">Started:</span>
                  <span className="meta-value">
                    {new Date(session.start_time).toLocaleString()}
                  </span>
                </span>
              </div>
            </button>
          ))
        )}
      </div>
    </div>
  )
}

export default Sessions
