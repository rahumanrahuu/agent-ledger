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

function App() {
  const [currentView, setCurrentView] = useState('overview')
  const [inspectorOpen, setInspectorOpen] = useState(false)
  const [inspectorData, setInspectorData] = useState(null)
  const [overview, setOverview] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

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
      <Sidebar currentView={currentView} onViewChange={setCurrentView} />
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
    </div>
  )
}

export default App
