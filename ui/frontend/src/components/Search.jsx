import { useState } from 'react'
import {
  FiX,
  FiTerminal,
  FiCheckSquare,
  FiCompass,
  FiAlertTriangle,
  FiShield,
  FiBookmark,
  FiFileText,
} from 'react-icons/fi'
import './Search.css'

function Search({ onSelect, onClose }) {
  const [query, setQuery] = useState('')
  const [results, setResults] = useState([])
  const [loading, setLoading] = useState(false)
  const [searched, setSearched] = useState(false)

  const handleSearch = async (e) => {
    e.preventDefault()
    if (!query.trim()) return

    setLoading(true)
    setSearched(true)
    try {
      const response = await fetch(`/api/search?q=${encodeURIComponent(query)}`)
      if (!response.ok) throw new Error('Search failed')
      const data = await response.json()
      setResults(data.results || [])
    } catch (err) {
      console.error(err)
      setResults([])
    } finally {
      setLoading(false)
    }
  }

  const typeIcons = {
    session: <FiTerminal />,
    decision: <FiCheckSquare />,
    discovery: <FiCompass />,
    failure: <FiAlertTriangle />,
    constraint: <FiShield />,
    checkpoint: <FiBookmark />,
  }

  return (
    <div className="search-overlay" onClick={onClose}>
      <div className="search-modal" onClick={(e) => e.stopPropagation()}>
        <div className="search-header">
          <h2>Search</h2>
          <button className="close-btn" onClick={onClose}>
            <FiX />
          </button>
        </div>

        <form onSubmit={handleSearch} className="search-form">
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search sessions, decisions, discoveries..."
            autoFocus
            className="search-input"
          />
          <button type="submit" className="search-btn">Search</button>
        </form>

        <div className="search-results">
          {loading && <p className="search-status">Searching...</p>}
          {searched && results.length === 0 && !loading && (
            <p className="search-status">No results found</p>
          )}
          {results.map((result) => (
            <button
              key={`${result.type}-${result.id}`}
              className="search-result"
              onClick={() => {
                onSelect({ type: result.type, id: result.id, ...result })
                onClose()
              }}
            >
              <span className="result-icon">{typeIcons[result.type]}</span>
              <div className="result-content">
                <div className="result-title">
                  {result.title}
                  <span className="result-type">{result.type}</span>
                </div>
                <p className="result-excerpt">{result.excerpt}</p>
              </div>
            </button>
          ))}
        </div>
      </div>
    </div>
  )
}

export default Search
