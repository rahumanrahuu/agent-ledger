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
import ProjectView from './views/ProjectView'

function App() {
  const [currentView, setCurrentView] = useState('overview')
  const [selectedItem, setSelectedItem] = useState(null)
  const [overview, setOverview] = useState(null)
  const [loading, setLoading] = useState(true)
  const [revision, setRevision] = useState(0)
  const [liveStatus, setLiveStatus] = useState('connecting')

  useEffect(() => {
    let socket
    let retry
    let stopped = false
    const connect = () => {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      socket = new WebSocket(`${protocol}//${window.location.host}/api/live`)
      socket.onopen = () => setLiveStatus('connected')
      socket.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data)
          if (message.type === 'ledger.updated') setRevision(value => value + 1)
        } catch {
          // API fetches remain authoritative if a notification is malformed.
        }
      }
      socket.onerror = () => socket.close()
      socket.onclose = () => {
        if (!stopped) {
          setLiveStatus('reconnecting')
          retry = window.setTimeout(connect, 1500)
        }
      }
    }
    connect()
    return () => {
      stopped = true
      window.clearTimeout(retry)
      socket?.close()
    }
  }, [])

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
  }, [revision])

  const renderView = () => {
    switch (currentView) {
      case 'overview':
        return <Overview data={overview} onSelect={setSelectedItem} onNavigate={setCurrentView} liveStatus={liveStatus} revision={revision} />
      case 'sessions':
        return <Sessions onSelect={setSelectedItem} revision={revision} />
      case 'timeline':
        return <Timeline onSelect={setSelectedItem} revision={revision} />
      case 'decisions':
        return <Decisions onSelect={setSelectedItem} revision={revision} />
      case 'discoveries':
        return <Discoveries onSelect={setSelectedItem} revision={revision} />
      case 'checkpoints':
        return <Checkpoints onSelect={setSelectedItem} revision={revision} />
      case 'knowledge-graph':
        return <KnowledgeGraph revision={revision} onNodeClick={setSelectedItem} onNavigate={(view, node) => {
          setCurrentView(view)
          setSelectedItem({ type: node.type, ...node })
        }} />
      case 'files':
        return <ProjectView kind="files" data={overview} onSelect={setSelectedItem} />
      case 'metadata':
        return <ProjectView kind="metadata" data={overview} onSelect={setSelectedItem} />
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
