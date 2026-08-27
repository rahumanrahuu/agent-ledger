import React, { useState, useEffect } from 'react'
import './BitbucketMemoryPanel.css'

export default function BitbucketMemoryPanel() {
  const [memories, setMemories] = useState([])
  const [loading, setLoading] = useState(false)
  const [searchQuery, setSearchQuery] = useState('')
  const [filterType, setFilterType] = useState('all')

  useEffect(() => {
    fetchMemories()
  }, [filterType])

  const fetchMemories = async () => {
    setLoading(true)
    try {
      const response = await fetch(`/api/memories?type=${filterType}&limit=20`)
      const data = await response.json()
      setMemories(data || [])
    } catch (error) {
      console.error('Failed to fetch memories:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleSearch = async (e) => {
    e.preventDefault()
    if (!searchQuery.trim()) {
      fetchMemories()
      return
    }

    setLoading(true)
    try {
      const response = await fetch(`/api/search?q=${encodeURIComponent(searchQuery)}&limit=20`)
      const data = await response.json()
      setMemories(data?.results || [])
    } catch (error) {
      console.error('Search failed:', error)
    } finally {
      setLoading(false)
    }
  }

  const getTypeColor = (type) => {
    const colors = {
      decision: '#0052cc',
      discovery: '#216e4e',
      constraint: '#974f0c',
      failure: '#ae2a19',
      default: '#626f86',
    }
    return colors[type] || colors.default
  }

  const getTypeLabel = (type) => {
    const labels = {
      decision: 'Decision',
      discovery: 'Discovery',
      constraint: 'Constraint',
      failure: 'Failure',
    }
    return labels[type] || type
  }

  const formatDate = (date) => {
    return new Date(date).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    })
  }

  return (
    <div className="bb-memory-panel">
      {/* Header */}
      <div className="bb-panel-header">
        <h1>Memories</h1>
        <p className="bb-subtitle">Knowledge base and session history</p>
      </div>

      {/* Search and Filters */}
      <div className="bb-search-section">
        <form onSubmit={handleSearch} className="bb-search-form">
          <div className="bb-search-input-wrapper">
            <svg width="16" height="16" viewBox="0 0 16 16" className="bb-search-icon">
              <circle cx="7" cy="7" r="5" fill="none" stroke="currentColor" strokeWidth="1.5" />
              <path d="M11 11l3 3" stroke="currentColor" strokeWidth="1.5" />
            </svg>
            <input
              type="text"
              placeholder="Search memories..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="bb-search-input"
            />
          </div>
          <button type="submit" className="bb-btn bb-btn-primary">
            Search
          </button>
        </form>

        <div className="bb-filters">
          <div className="bb-filter-group">
            <label htmlFor="type-filter" className="bb-filter-label">
              Type:
            </label>
            <select
              id="type-filter"
              value={filterType}
              onChange={(e) => setFilterType(e.target.value)}
              className="bb-select"
            >
              <option value="all">All Types</option>
              <option value="decision">Decisions</option>
              <option value="discovery">Discoveries</option>
              <option value="constraint">Constraints</option>
              <option value="failure">Failures</option>
            </select>
          </div>

          <div className="bb-stats">
            <span className="bb-stat">
              <strong>{memories.length}</strong> memories
            </span>
          </div>
        </div>
      </div>

      {/* Memory List */}
      <div className="bb-memory-list">
        {loading ? (
          <div className="bb-loading">
            <div className="bb-spinner"></div>
            <p>Loading memories...</p>
          </div>
        ) : memories.length === 0 ? (
          <div className="bb-empty-state">
            <svg width="48" height="48" viewBox="0 0 48 48" fill="currentColor" opacity="0.3">
              <path d="M8 8h32v32H8V8zm2 2v28h28V10H10z" />
            </svg>
            <h3>No memories found</h3>
            <p>Create or search for memories to get started</p>
          </div>
        ) : (
          <table className="bb-table">
            <thead>
              <tr>
                <th>Title</th>
                <th>Type</th>
                <th>Created</th>
                <th>Session</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {memories.map((memory) => (
                <tr key={memory.id} className="bb-table-row">
                  <td className="bb-table-cell-title">
                    <a href={`#memory/${memory.id}`} className="bb-memory-link">
                      {memory.title}
                    </a>
                  </td>
                  <td className="bb-table-cell-type">
                    <span
                      className="bb-badge"
                      style={{ borderColor: getTypeColor(memory.type) }}
                    >
                      {getTypeLabel(memory.type)}
                    </span>
                  </td>
                  <td className="bb-table-cell-date">
                    {formatDate(memory.created_at)}
                  </td>
                  <td className="bb-table-cell-session">
                    <code className="bb-code">{memory.session_id?.slice(0, 8)}</code>
                  </td>
                  <td className="bb-table-cell-actions">
                    <button
                      className="bb-icon-btn-small"
                      aria-label="View"
                      title="View details"
                    >
                      <svg width="16" height="16" viewBox="0 0 16 16">
                        <path
                          d="M8 3C4.5 3 1.7 5.3 1 8c.7 2.7 3.5 5 7 5s6.3-2.3 7-5c-.7-2.7-3.5-5-7-5zm0 8c-1.66 0-3-1.34-3-3s1.34-3 3-3 3 1.34 3 3-1.34 3-3 3z"
                          fill="currentColor"
                        />
                      </svg>
                    </button>
                    <button
                      className="bb-icon-btn-small"
                      aria-label="Delete"
                      title="Delete"
                    >
                      <svg width="16" height="16" viewBox="0 0 16 16">
                        <path
                          d="M2 2v12h12V2H2zm3 2h6v8H5V4z"
                          fill="currentColor"
                        />
                      </svg>
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  )
}
