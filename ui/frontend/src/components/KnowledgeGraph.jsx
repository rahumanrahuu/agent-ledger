import { useState, useEffect, useRef } from 'react'
import './KnowledgeGraph.css'

function KnowledgeGraph({ onNodeClick }) {
  const [graph, setGraph] = useState(null)
  const [loading, setLoading] = useState(true)
  const [selectedNode, setSelectedNode] = useState(null)
  const [transform, setTransform] = useState({ x: 0, y: 0, scale: 1 })
  const svgRef = useRef(null)
  const dragStart = useRef(null)

  useEffect(() => {
    const fetchGraph = async () => {
      try {
        const response = await fetch('/api/graph')
        if (!response.ok) throw new Error('Failed to fetch graph')
        const data = await response.json()
        setGraph(data)
      } catch (err) {
        console.error(err)
      } finally {
        setLoading(false)
      }
    }

    fetchGraph()
  }, [])

  const handleWheel = (e) => {
    e.preventDefault()
    const scale = transform.scale * (e.deltaY > 0 ? 0.9 : 1.1)
    setTransform({ ...transform, scale: Math.max(0.5, Math.min(3, scale)) })
  }

  const handleMouseDown = (e) => {
    dragStart.current = { x: e.clientX - transform.x, y: e.clientY - transform.y }
  }

  const handleMouseMove = (e) => {
    if (!dragStart.current) return
    setTransform({
      ...transform,
      x: e.clientX - dragStart.current.x,
      y: e.clientY - dragStart.current.y,
    })
  }

  const handleMouseUp = () => {
    dragStart.current = null
  }

  if (loading) return <div className="graph-container">Loading graph...</div>
  if (!graph || !graph.nodes || graph.nodes.length === 0)
    return <div className="graph-container">No data for graph visualization</div>

  // Simple force-directed layout
  const width = 800
  const height = 600

  // Position nodes using a simple grid layout based on type
  const positions = {}
  const typeGroups = {}
  (graph.nodes || []).forEach((node) => {
    if (!typeGroups[node.type]) typeGroups[node.type] = []
    typeGroups[node.type].push(node)
  })

  let yOffset = 100
  Object.entries(typeGroups).forEach(([type, nodes]) => {
    const xStart = width / (Object.keys(typeGroups).length + 1)
    const typeIndex = Object.keys(typeGroups).indexOf(type)
    const x = ((typeIndex + 1) * width) / (Object.keys(typeGroups).length + 1)

    nodes.forEach((node, idx) => {
      const y = yOffset + idx * 80
      positions[node.id] = { x, y }
    })
  })

  const typeColors = {
    session: '#0071e3',
    decision: '#34c759',
    discovery: '#ff9500',
    checkpoint: '#a0a0a0',
    failure: '#ff3b30',
    constraint: '#ffcc00',
  }

  const typeIcons = {
    session: '●',
    decision: '◈',
    discovery: '▲',
    checkpoint: '■',
    failure: '✕',
    constraint: '◆',
  }

  return (
    <div
      className="graph-container"
      onWheel={handleWheel}
      onMouseDown={handleMouseDown}
      onMouseMove={handleMouseMove}
      onMouseUp={handleMouseUp}
      onMouseLeave={handleMouseUp}
    >
      <svg ref={svgRef} width={width} height={height} className="graph-svg">
        <defs>
          <marker
            id="arrowhead"
            markerWidth="10"
            markerHeight="10"
            refX="9"
            refY="3"
            orient="auto"
          >
            <polygon points="0 0, 10 3, 0 6" fill="#e5e5e7" />
          </marker>
        </defs>

        <g style={{ transform: `translate(${transform.x}px, ${transform.y}px) scale(${transform.scale})` }}>
          {/* Draw edges */}
          {graph.edges.map((edge, idx) => {
            const from = positions[edge.source]
            const to = positions[edge.target]
            if (!from || !to) return null

            return (
              <line
                key={`edge-${idx}`}
                x1={from.x}
                y1={from.y}
                x2={to.x}
                y2={to.y}
                stroke="#e5e5e7"
                strokeWidth="1"
                markerEnd="url(#arrowhead)"
              />
            )
          })}

          {/* Draw nodes */}
          {graph.nodes.map((node) => {
            const pos = positions[node.id]
            if (!pos) return null

            const isSelected = selectedNode?.id === node.id
            const radius = isSelected ? 28 : 20

            return (
              <g
                key={node.id}
                onClick={() => {
                  setSelectedNode(node)
                  onNodeClick?.(node)
                }}
              >
                {/* Node circle */}
                <circle
                  cx={pos.x}
                  cy={pos.y}
                  r={radius}
                  fill={typeColors[node.type] || '#999999'}
                  opacity={isSelected ? 1 : 0.8}
                  stroke={isSelected ? '#0071e3' : 'white'}
                  strokeWidth={isSelected ? 2 : 1}
                  style={{ cursor: 'pointer', transition: 'all 0.2s' }}
                />

                {/* Node icon */}
                <text
                  x={pos.x}
                  y={pos.y}
                  textAnchor="middle"
                  dy="0.3em"
                  fill="white"
                  fontSize="14"
                  fontWeight="bold"
                  pointerEvents="none"
                >
                  {typeIcons[node.type]}
                </text>

                {/* Node label */}
                {isSelected && (
                  <text
                    x={pos.x}
                    y={pos.y + radius + 20}
                    textAnchor="middle"
                    fontSize="12"
                    fill="#333333"
                    pointerEvents="none"
                  >
                    {node.label?.substring(0, 15)}
                  </text>
                )}
              </g>
            )
          })}
        </g>
      </svg>

      {selectedNode && (
        <div className="graph-info">
          <h4>{selectedNode.label}</h4>
          <span className="info-type">{selectedNode.type}</span>
          <p className="info-hint">Click to view details</p>
        </div>
      )}

      <div className="graph-controls">
        <span className="control-hint">Scroll to zoom • Drag to pan</span>
      </div>
    </div>
  )
}

export default KnowledgeGraph
