import { useState, useEffect } from 'react'
import './App.css'
import Sidebar from './components/Sidebar'
import Overview from './views/Overview'
import Sessions from './views/Sessions'
import Timeline from './views/Timeline'
import Decisions from './views/Decisions'
import Discoveries from './views/Discoveries'
import Checkpoints from './views/Checkpoints'
import KnowledgeGraph from './components/KnowledgeGraph'
import Inspector from './components/Inspector'

function App() {
  const [currentView, setCurrentView] = useState('overview')
  const [selectedItem, setSelectedItem] = useState(null)
  const [overview, setOverview] = useState(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const fetchOverview = async () => {
      try {
        const response = await fetch('/api/overview')
        if (!response.ok) throw new Error('Failed to fetch overview')
        const data = await response.json()
        setOverview(data)
      } catch (err) {
        console.error('Failed to load overview:', err)
      } finally {
        setLoading(false)
      }
    }

    fetchOverview()
  }, [])

  const renderView = () => {
    switch (currentView) {
      case 'overview':
        return <Overview data={overview} onSelect={setSelectedItem} />
      case 'sessions':
        return <Sessions onSelect={setSelectedItem} />
      case 'timeline':
        return <Timeline onSelect={setSelectedItem} />
      case 'decisions':
        return <Decisions onSelect={setSelectedItem} />
      case 'discoveries':
        return <Discoveries onSelect={setSelectedItem} />
      case 'checkpoints':
        return <Checkpoints onSelect={setSelectedItem} />
      case 'knowledge-graph':
        return <KnowledgeGraph onNodeClick={setSelectedItem} />
      default:
        return <Overview data={overview} onSelect={setSelectedItem} />
    }
  }

  if (loading) {
    return (
      <div className="app">
        <div className="loading">Loading Agent Ledger...</div>
      </div>
    )
  }

  return (
    <div className="app">
      <Sidebar currentView={currentView} onViewChange={setCurrentView} />
      <main className="main-area">
        <div className="center-column">{renderView()}</div>
        {selectedItem && (
          <Inspector
            item={selectedItem}
            onClose={() => setSelectedItem(null)}
          />
        )}
      </main>
    </div>
  )
}

export default App
