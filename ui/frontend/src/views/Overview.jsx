import KnowledgeGraph from '../components/KnowledgeGraph'
import './Overview.css'

function Overview({ data, onSelect }) {
  if (!data) return null

  const metrics = [
    { label: 'Sessions', value: data.session_count, color: 'blue' },
    { label: 'Decisions', value: data.decision_count, color: 'green' },
    { label: 'Discoveries', value: data.discovery_count, color: 'yellow' },
    { label: 'Checkpoints', value: data.checkpoint_count, color: 'gray' },
  ]

  return (
    <div className="overview">
      <header className="overview-header">
        <div>
          <h1>{data.project_name}</h1>
          <p className="text-secondary">{data.repository_root}</p>
        </div>
      </header>

      <section className="overview-section">
        <h2>Project Status</h2>
        <div className="status-grid">
          <div className="status-card">
            <span className="label">Branch</span>
            <span className="value font-mono">{data.current_branch}</span>
          </div>
          <div className="status-card">
            <span className="label">Commit</span>
            <span className="value font-mono">{data.current_commit?.substring(0, 7)}</span>
          </div>
          <div className="status-card">
            <span className="label">Version</span>
            <span className="value">{data.version}</span>
          </div>
          {data.last_activity_time && (
            <div className="status-card">
              <span className="label">Last Activity</span>
              <span className="value">
                {new Date(data.last_activity_time).toLocaleDateString()}
              </span>
            </div>
          )}
        </div>
      </section>

      <section className="overview-section">
        <h2>Metrics</h2>
        <div className="metrics-grid">
          {metrics.map((metric) => (
            <button
              key={metric.label}
              className={`metric-card metric-${metric.color}`}
              onClick={() => onSelect({ type: 'metric', name: metric.label })}
            >
              <div className="metric-value">{metric.value}</div>
              <div className="metric-label">{metric.label}</div>
            </button>
          ))}
        </div>
      </section>

      <section className="overview-section">
        <h2>Knowledge Graph</h2>
        <KnowledgeGraph onNodeClick={onSelect} />
      </section>
    </div>
  )
}

export default Overview
