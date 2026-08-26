import { useState, useEffect } from 'react'
import {
  FiCheckSquare,
  FiCompass,
  FiAlertTriangle,
  FiShield,
  FiBookmark,
} from 'react-icons/fi'
import Markdown from '../components/Markdown'
import './Timeline.css'

function Timeline({ onSelect }) {
  const [events, setEvents] = useState([])
  const [loading, setLoading] = useState(true)
  const [filter, setFilter] = useState('')
  const [selected, setSelected] = useState(null)

  useEffect(() => {
    const fetchEvents = async () => {
      try {
        const response = await fetch('/api/events')
        if (!response.ok) throw new Error('Failed to fetch events')
        const data = await response.json()
        setEvents(data.events || [])
      } catch (err) {
        console.error(err)
      } finally {
        setLoading(false)
      }
    }

    fetchEvents()
  }, [])

  const filteredEvents = filter
    ? events.filter((e) => e.type === filter)
    : events

  const typeIcons = {
    decision: <FiCheckSquare />,
    discovery: <FiCompass />,
    failure: <FiAlertTriangle />,
    constraint: <FiShield />,
    checkpoint: <FiBookmark />,
  }

  const typeColors = {
    decision: 'type-decision',
    discovery: 'type-discovery',
    failure: 'type-failure',
    constraint: 'type-constraint',
    checkpoint: 'type-checkpoint',
  }

  if (loading) return <div className="view-container">Loading timeline...</div>

  return (
    <div className="view-container">
      <header className="view-header">
        <h1>Timeline</h1>
        <p className="text-secondary">Chronological stream of all events</p>
      </header>

      <div className="timeline-controls">
        <button
          className={`filter-btn ${filter === '' ? 'active' : ''}`}
          onClick={() => setFilter('')}
        >
          All
        </button>
        <button
          className={`filter-btn ${filter === 'decision' ? 'active' : ''}`}
          onClick={() => setFilter('decision')}
        >
          Decisions
        </button>
        <button
          className={`filter-btn ${filter === 'discovery' ? 'active' : ''}`}
          onClick={() => setFilter('discovery')}
        >
          Discoveries
        </button>
        <button
          className={`filter-btn ${filter === 'failure' ? 'active' : ''}`}
          onClick={() => setFilter('failure')}
        >
          Failures
        </button>
        <button
          className={`filter-btn ${filter === 'constraint' ? 'active' : ''}`}
          onClick={() => setFilter('constraint')}
        >
          Constraints
        </button>
      </div>

      {filteredEvents.length === 0 ? (
        <div className="empty-state">
          <p>No events found</p>
        </div>
      ) : (
        <div className="timeline-wrapper">
          <div className="timeline">
            {filteredEvents.map((event, idx) => (
              <div key={event.id} className="timeline-item">
                <div className={`timeline-dot ${typeColors[event.type]}`}>
                  {typeIcons[event.type]}
                </div>
                <button
                  className={`timeline-card ${selected?.id === event.id ? 'active' : ''}`}
                  onClick={() => {
                    setSelected(event)
                    onSelect({ type: event.type, ...event })
                  }}
                >
                  <div className="timeline-card-header">
                    <span className="event-type">{event.type}</span>
                    <span className="event-date">
                      {new Date(event.timestamp).toLocaleDateString()}
                    </span>
                  </div>
                  <h3>{event.title}</h3>
                  <p className="event-excerpt">
                    {event.content?.substring(0, 100) || 'No content'}...
                  </p>
                </button>
              </div>
            ))}
          </div>

          {selected && (
            <div className="timeline-detail">
              <div className="detail-header">
                <h2>{selected.title}</h2>
                <span className={`detail-type ${typeColors[selected.type]}`}>
                  {selected.type}
                </span>
              </div>
              <p className="detail-date">
                {new Date(selected.timestamp).toLocaleString()}
              </p>
              <Markdown content={selected.content} />
              <div className="detail-path">
                <span className="text-secondary text-sm">Path: {selected.path}</span>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  )
}

export default Timeline
