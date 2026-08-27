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
    fetch('/api/overview')
      .then(r => r.json())
      .then(data => {
        setOverview(data)
        setLoading(false)
      })
  }, [revision])

  const renderContent = () => {
    switch (currentView) {
      case 'overview':
        return <Overview data={overview} />
      case 'sessions':
        return <Sessions />
      case 'timeline':
        return <Timeline />
      case 'memories':
        return <BitbucketMemoryPanel />
      case 'graph':
        return <KnowledgeGraph />
      default:
        return <Overview data={overview} />
    }
  }

  return (
    <BitbucketLayout>
      <div className="bb-content-wrapper">
        <main className="bb-content-main">
          {renderContent()}
        </main>
        <aside className="bb-content-inspector">
          <Inspector selectedItem={selectedItem} liveStatus={liveStatus} />
        </aside>
      </div>
    </BitbucketLayout>
  )
}

export default App
