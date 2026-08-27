import { useState, useEffect } from 'react'
import './App.css'
import BitbucketLayout from './components/BitbucketLayout'
import BitbucketMemoryPanel from './components/BitbucketMemoryPanel'
import Overview from './views/Overview'
import Sessions from './views/Sessions'
import Timeline from './views/Timeline'
import KnowledgeGraph from './components/KnowledgeGraph'
import Inspector from './components/Inspector'

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
      case 'memories':
        return <BitbucketMemoryPanel />
      case 'knowledge-graph':
        return <KnowledgeGraph revision={revision} onNodeClick={setSelectedItem} />
      default:
        return <Overview data={overview} onSelect={setSelectedItem} />
    }
  }

  if (loading) {
    return (
      <BitbucketLayout>
        <div className="bb-loading-page">
          <div className="bb-spinner"></div>
          <p>Loading Agent Ledger...</p>
        </div>
      </BitbucketLayout>
    )
  }

  return (
    <BitbucketLayout onNavigate={setCurrentView} currentView={currentView}>
      <div className="bb-app-content">
        <div className="bb-view-container">
          {renderView()}
        </div>
        {selectedItem && (
          <div className="bb-inspector-container">
            <Inspector
              item={selectedItem}
              onClose={() => setSelectedItem(null)}
            />
          </div>
        )}
      </div>
    </BitbucketLayout>
  )
}

export default App
