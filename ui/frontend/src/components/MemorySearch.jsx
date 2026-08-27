import { useState, useEffect } from 'react'
import { FiSearch, FiX, FiFilter, FiTrendingUp } from 'react-icons/fi'
import './MemorySearch.css'

function MemorySearch({ onSelect, isOpen, onClose }) {
  const [query, setQuery] = useState('')
  const [results, setResults] = useState([])
  const [loading, setLoading] = useState(false)
  const [filters, setFilters] = useState({
    type: 'all',
    timeRange: '30d',
    minRelevance: 0.6,
  })
  const [searchHistory, setSearchHistory] = useState([])

  useEffect(() => {
    const performSearch = async () => {
      if (!query.trim()) {
        setResults([])
        return
      }

      setLoading(true)
      try {
        const params = new URLSearchParams({
          q: query,
          type: filters.type !== 'all' ? filters.type : '',
          threshold: filters.minRelevance,
        })
        const response = await fetch(`/api/search?${params}`)
        if (!response.ok) throw new Error('Search failed')

        const data = await response.json()
        setResults(data.results || [])

        // Add to history
        if (!searchHistory.includes(query)) {
          setSearchHistory([query, ...searchHistory.slice(0, 9)])
        }
      } catch (err) {
        console.error('Search error:', err)
        setResults([])
      } finally {
        setLoading(false)
      }
    }

    const timer = setTimeout(performSearch, 300)
    return () => clearTimeout(timer)
  }, [query, filters])

  const typeColors = {
    decision: 'var(--color-blue)',
    discovery: 'var(--color-green)',
    constraint: 'var(--color-yellow)',
    failure: 'var(--color-red)',
  }

  if (!isOpen) return null

  return (
    <div className="memory-search-overlay" onClick={onClose}>
      <div className="memory-search-panel" onClick={(e) => e.stopPropagation()}>
        <div className="search-header">
          <h2>🧠 Memory Search</h2>
          <button className="close-btn" onClick={onClose}>
            <FiX />
          </button>
        </div>

        <div className="search-input-group">
          <FiSearch className="search-icon" />
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search decisions, discoveries, constraints..."
            autoFocus
            className="search-input"
          />
        </div>

        <div className="search-filters">
          <select
            value={filters.type}
            onChange={(e) => setFilters({...filters, type: e.target.value})}
            className="filter-select"
          >
            <option value="all">All Types</option>
            <option value="decision">Decisions</option>
            <option value="discovery">Discoveries</option>
            <option value="constraint">Constraints</option>
            <option value="failure">Failures</option>
          </select>

          <select
            value={filters.timeRange}
            onChange={(e) => setFilters({...filters, timeRange: e.target.value})}
            className="filter-select"
          >
            <option value="7d">Last 7 days</option>
            <option value="30d">Last 30 days</option>
            <option value="90d">Last 90 days</option>
            <option value="all">All time</option>
          </select>

          <input
            type="range"
            min="0.5"
            max="1"
            step="0.05"
            value={filters.minRelevance}
            onChange={(e) => setFilters({...filters, minRelevance: parseFloat(e.target.value)})}
            className="relevance-slider"
            title="Minimum relevance score"
          />
          <span className="relevance-value">{(filters.minRelevance * 100).toFixed(0)}%</span>
        </div>

        <div className="search-results-container">
          {loading && (
            <div className="search-loading">
              <div className="spinner" />
              Searching memories...
            </div>
          )}

          {!query.trim() && searchHistory.length > 0 && (
            <div className="search-history">
              <p className="history-label">Recent Searches:</p>
              <div className="history-items">
                {searchHistory.map((prev, i) => (
                  <button
                    key={i}
                    className="history-item"
                    onClick={() => setQuery(prev)}
                  >
                    {prev}
                  </button>
                ))}
              </div>
            </div>
          )}

          {!loading && results.length === 0 && query.trim() && (
            <div className="search-empty">
              <p>No memories found matching "{query}"</p>
              <p className="hint">Try different keywords or adjust filters</p>
            </div>
          )}

          {!loading && results.length > 0 && (
            <div className="search-results">
              <p className="results-count">{results.length} results</p>
              {results.map((result) => (
                <button
                  key={result.id}
                  className="result-item"
                  onClick={() => onSelect(result)}
                  style={{ borderLeftColor: typeColors[result.type] || 'var(--color-border)' }}
                >
                  <div className="result-header">
                    <span className="result-type" style={{ color: typeColors[result.type] }}>
                      {result.type.toUpperCase()}
                    </span>
                    <span className="result-score">
                      <FiTrendingUp size={14} /> {(result.score * 100).toFixed(0)}%
                    </span>
                  </div>
                  <h4 className="result-title">{result.title}</h4>
                  <p className="result-excerpt">{result.excerpt || result.content?.substring(0, 120)}</p>
                  <div className="result-meta">
                    <span className="meta-date">{new Date(result.created_at).toLocaleDateString()}</span>
                    {result.session_id && <span className="meta-session">Session: {result.session_id.substring(0, 8)}</span>}
                  </div>
                </button>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

export default MemorySearch
