import { useState, useEffect } from 'react'
import './KnowledgeGraph.css'

function KnowledgeGraph({ onNodeClick }) {
  const [graph, setGraph] = useState(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const fetchGraph = async () => {
      try {
        const response = await fetch('/api/graph')
        if (!response.ok) throw new Error('Failed to fetch graph')
        const data = await response.json()
        setGraph(data)
      } catch (err) {
        console.error('Failed to fetch graph:', err)
      } finally {
        setLoading(false)
      }
    }

    fetchGraph()
  }, [])

  if (loading) {
    return <div className="graph-container">Loading graph...</div>
  }

  if (!graph || !graph.nodes || graph.nodes.length === 0) {
    return <div className="graph-container">No graph data available</div>
  }

  return (
    <div className="graph-container">
      <div style={{ width: '100%', height: '100%', display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', color: 'var(--color-text-secondary)' }}>
        <div style={{ textAlign: 'center' }}>
          <p style={{ fontSize: '14px', margin: '0 0 10px 0' }}>Knowledge Graph</p>
          <p style={{ fontSize: '12px', color: 'var(--color-text-tertiary)', margin: 0 }}>
            Nodes: {graph.nodes?.length || 0} | Edges: {graph.edges?.length || 0}
          </p>
        </div>
      </div>
    </div>
  )
}

export default KnowledgeGraph
