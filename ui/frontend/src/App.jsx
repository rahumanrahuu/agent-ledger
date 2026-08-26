import { useState, useEffect } from 'react'
import './App.css'
import Sidebar from './components/Sidebar'
import Overview from './views/Overview'
import Sessions from './views/Sessions'
import Decisions from './views/Decisions'
import Discoveries from './views/Discoveries'
import Checkpoints from './views/Checkpoints'
import Timeline from './views/Timeline'
import Inspector from './components/Inspector'
import Search from './components/Search'

function App() {
  const [currentView, setCurrentView] = useState('overview')
  const [inspectorOpen, setInspectorOpen] = useState(false)
  const [inspectorData, setInspectorData] = useState(null)
  const [overview, setOverview] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [searchOpen, setSearchOpen] = useState(false)

  useEffect(() => {
    const fetchOverview = async () => {
      try {
        const response = await fetch('/api/overview')
        if (!response.ok) throw new Error('Failed to fetch overview')
        const data = await response.json()
        setOverview(data)
      } catch (err) {
        setError(err.message)
      } finally {
        setLoading(false)
      }
    }

    fetchOverview()
  }, [])

  // Handle keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (e) => {
      // Cmd+K or Ctrl+K for search
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault()
        setSearchOpen(!searchOpen)
      }
      // Escape to close search
      if (e.key === 'Escape' && searchOpen) {
        setSearchOpen(false)
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [searchOpen])

  const handleSelectItem = (data) => {
    setInspectorData(data)
    setInspectorOpen(true)
  }

  const renderView = () => {
    switch (currentView) {
      case 'overview':
        return <Overview data={overview} onSelect={handleSelectItem} />
      case 'sessions':
        return <Sessions onSelect={handleSelectItem} />
      case 'decisions':
        return <Decisions onSelect={handleSelectItem} />
      case 'discoveries':
        return <Discoveries onSelect={handleSelectItem} />
      case 'checkpoints':
        return <Checkpoints onSelect={handleSelectItem} />
      case 'timeline':
        return <Timeline onSelect={handleSelectItem} />
      default:
        return <Overview data={overview} onSelect={handleSelectItem} />
    }
  }

  return (
    <div className="app">
      <Sidebar
        currentView={currentView}
        onViewChange={setCurrentView}
        onSearchClick={() => setSearchOpen(true)}
      />
      <main className="main-content">
        {loading ? (
          <div className="loading-state">Loading...</div>
        ) : error ? (
          <div className="error-state">Error: {error}</div>
        ) : (
          renderView()
        )}
      </main>
      {inspectorOpen && (
        <Inspector
          data={inspectorData}
          onClose={() => setInspectorOpen(false)}
        />
      )}
      {searchOpen && (
        <Search
          onSelect={handleSelectItem}
          onClose={() => setSearchOpen(false)}
        />
      )}
    </div>
  )
}

export default App
