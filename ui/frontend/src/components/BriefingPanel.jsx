import { useState, useEffect } from 'react'
import { FiChevronDown, FiChevronUp, FiX, FiAlertCircle, FiZap, FiCheck } from 'react-icons/fi'
import './BriefingPanel.css'

function BriefingPanel({ task, onDismiss }) {
  const [briefing, setBriefing] = useState(null)
  const [loading, setLoading] = useState(true)
  const [expanded, setExpanded] = useState({
    techStack: true,
    decisions: true,
    constraints: false,
    risks: false,
  })

  useEffect(() => {
    const fetchBriefing = async () => {
      try {
        const response = await fetch(`/api/briefing?task=${encodeURIComponent(task)}`)
        if (!response.ok) throw new Error('Failed to fetch briefing')
        const data = await response.json()
        setBriefing(data)
      } catch (err) {
        console.error('Briefing error:', err)
      } finally {
        setLoading(false)
      }
    }

    fetchBriefing()
  }, [task])

  if (!briefing && !loading) return null

  const toggleSection = (section) => {
    setExpanded((prev) => ({...prev, [section]: !prev[section]}))
  }

  return (
    <div className="briefing-panel">
      <div className="briefing-header">
        <div>
          <h3>🧠 Project Briefing</h3>
          <p className="briefing-task">{task}</p>
        </div>
        <button className="dismiss-btn" onClick={onDismiss}>
          <FiX />
        </button>
      </div>

      {loading ? (
        <div className="briefing-loading">
          <div className="spinner" />
          Generating briefing...
        </div>
      ) : (
        <div className="briefing-content">
          {/* Estimated Duration */}
          <div className="briefing-quick-stats">
            <div className="stat">
              <span className="stat-label">Estimated Duration</span>
              <span className="stat-value">{briefing.estimated_duration || '45-60m'}</span>
            </div>
            <div className="stat">
              <span className="stat-label">Estimated Cost</span>
              <span className="stat-value">2-4k tokens</span>
            </div>
          </div>

          {/* Tech Stack */}
          {briefing.tech_stack && briefing.tech_stack.length > 0 && (
            <section className="briefing-section">
              <button
                className="section-header"
                onClick={() => toggleSection('techStack')}
              >
                <span className="section-title">📚 Tech Stack</span>
                {expanded.techStack ? <FiChevronUp /> : <FiChevronDown />}
              </button>
              {expanded.techStack && (
                <div className="section-content">
                  <ul className="tech-list">
                    {briefing.tech_stack.map((item, i) => (
                      <li key={i}>{item}</li>
                    ))}
                  </ul>
                </div>
              )}
            </section>
          )}

          {/* Architecture */}
          {briefing.architecture && (
            <section className="briefing-section">
              <button
                className="section-header"
                onClick={() => toggleSection('architecture')}
              >
                <span className="section-title">🏗️ Architecture</span>
                {expanded.architecture ? <FiChevronUp /> : <FiChevronDown />}
              </button>
              {expanded.architecture && (
                <div className="section-content">
                  <p className="architecture-text">{briefing.architecture}</p>
                </div>
              )}
            </section>
          )}

          {/* Constraints - Always show summary */}
          {briefing.constraints && briefing.constraints.length > 0 && (
            <section className="briefing-section constraints-section">
              <button
                className="section-header warning"
                onClick={() => toggleSection('constraints')}
              >
                <span className="section-title">🔴 Constraints ({briefing.constraints.length})</span>
                {expanded.constraints ? <FiChevronUp /> : <FiChevronDown />}
              </button>
              {expanded.constraints && (
                <div className="section-content">
                  <ul className="constraint-list">
                    {briefing.constraints.map((constraint, i) => (
                      <li key={i} className="constraint-item">
                        <FiAlertCircle size={16} />
                        <span>{constraint}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              )}
            </section>
          )}

          {/* Recent Decisions */}
          {briefing.decisions && briefing.decisions.length > 0 && (
            <section className="briefing-section">
              <button
                className="section-header"
                onClick={() => toggleSection('decisions')}
              >
                <span className="section-title">✓ Recent Decisions ({briefing.decisions.length})</span>
                {expanded.decisions ? <FiChevronUp /> : <FiChevronDown />}
              </button>
              {expanded.decisions && (
                <div className="section-content">
                  <ul className="decision-list">
                    {briefing.decisions.map((decision, i) => (
                      <li key={i} className="decision-item">
                        <FiCheck size={16} />
                        <span>{decision}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              )}
            </section>
          )}

          {/* Risks */}
          {briefing.risks && briefing.risks.length > 0 && (
            <section className="briefing-section">
              <button
                className="section-header alert"
                onClick={() => toggleSection('risks')}
              >
                <span className="section-title">⚠️ Known Risks ({briefing.risks.length})</span>
                {expanded.risks ? <FiChevronUp /> : <FiChevronDown />}
              </button>
              {expanded.risks && (
                <div className="section-content">
                  <ul className="risk-list">
                    {briefing.risks.map((risk, i) => (
                      <li key={i} className="risk-item">
                        <FiZap size={16} />
                        <span>{risk}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              )}
            </section>
          )}

          {/* Next Steps */}
          {briefing.next_steps && briefing.next_steps.length > 0 && (
            <section className="briefing-section">
              <button
                className="section-header"
                onClick={() => toggleSection('nextSteps')}
              >
                <span className="section-title">▶️ Next Steps</span>
                {expanded.nextSteps ? <FiChevronUp /> : <FiChevronDown />}
              </button>
              {expanded.nextSteps && (
                <div className="section-content">
                  <ol className="steps-list">
                    {briefing.next_steps.map((step, i) => (
                      <li key={i}>{step}</li>
                    ))}
                  </ol>
                </div>
              )}
            </section>
          )}

          {/* Actions */}
          <div className="briefing-actions">
            <button className="action-btn primary" onClick={onDismiss}>
              Start Working
            </button>
            <button className="action-btn secondary">
              View Full History
            </button>
          </div>
        </div>
      )}
    </div>
  )
}

export default BriefingPanel
