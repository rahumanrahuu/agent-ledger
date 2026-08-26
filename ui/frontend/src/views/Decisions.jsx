import { useState, useEffect } from 'react'
import Markdown from '../components/Markdown'
import './EventView.css'

function Decisions({ onSelect }) {
  const [events, setEvents] = useState([])
  const [loading, setLoading] = useState(true)
  const [selected, setSelected] = useState(null)

  useEffect(() => {
    const fetchDecisions = async () => {
      try {
        const response = await fetch('/api/events?type=decision')
        if (!response.ok) throw new Error('Failed to fetch decisions')
        const data = await response.json()
        setEvents(data.events || [])
      } catch (err) {
        console.error(err)
      } finally {
        setLoading(false)
      }
    }

    fetchDecisions()
  }, [])

  if (loading) return <div className="view-container">Loading decisions...</div>

  return (
    <div className="view-container">
      <header className="view-header">
        <h1>Decisions</h1>
        <p className="text-secondary">{events.length} decisions recorded</p>
      </header>

      {events.length === 0 ? (
        <div className="empty-state">
          <p>No decisions recorded yet</p>
        </div>
      ) : (
        <div className="event-grid">
          <div className="event-list">
            {events.map((event) => (
              <button
                key={event.id}
                className={`event-card ${selected?.id === event.id ? 'active' : ''}`}
                onClick={() => {
                  setSelected(event)
                  onSelect({ type: 'decision', ...event })
                }}
              >
                <div className="event-card-header">
                  <h3>{event.title}</h3>
                </div>
                <p className="event-preview">
                  {event.content?.substring(0, 80) || 'No content'}...
                </p>
                <span className="event-time">
                  {new Date(event.timestamp).toLocaleDateString()}
                </span>
              </button>
            ))}
          </div>

          {selected && (
            <div className="event-detail">
              <h2>{selected.title}</h2>
              <Markdown content={selected.content} />
              <div className="event-meta">
                <span className="text-secondary text-sm">
                  Path: <code>{selected.path}</code>
                </span>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  )
}

export default Decisions
