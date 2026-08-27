import React, { useState } from 'react'
import './BitbucketLayout.css'

const navItems = [
  ['overview', 'Overview', <path key="overview" d="M2 2h5v5H2V2zm7 0h5v5H9V2zM2 9h5v5H2V9zm7 0h5v5H9V9z" />],
  ['memories', 'Memories', <path key="memories" d="M3 1.5h8.5A1.5 1.5 0 0 1 13 3v11l-3-2-3 2-3-2-2 1.33V3A1.5 1.5 0 0 1 3.5 1.5H3Z" />],
  ['timeline', 'Timeline', <path key="timeline" d="M8 1.5a6.5 6.5 0 1 0 6.5 6.5A6.51 6.51 0 0 0 8 1.5Zm.75 3v3.19l2.1 2.1-1.06 1.06L7.25 8.31V4.5h1.5Z" />],
  ['sessions', 'Sessions', <path key="sessions" d="M8 1.5a3 3 0 1 1-3 3 3 3 0 0 1 3-3Zm0 7c3.04 0 5.5 1.57 5.5 3.5v2h-11v-2c0-1.93 2.46-3.5 5.5-3.5Z" />],
  ['knowledge-graph', 'Knowledge graph', <path key="graph" d="M3 1a2 2 0 1 1-1.28 3.54l2.15 4.3a2 2 0 0 1 2.98 1.1l3.28-.82a2 2 0 1 1 .36 1.46l-3.28.82A2 2 0 1 1 2.5 10c0-.32.08-.63.22-.9L.57 4.8A2 2 0 0 1 3 1Zm8.5 4a2 2 0 1 1 2-2 2 2 0 0 1-2 2Z" />],
]

export default function BitbucketLayout({ children, onNavigate, currentView }) {
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const navigate = (view) => { onNavigate?.(view); setSidebarOpen(false) }

  return <div className="bb-layout">
    <header className="bb-topnav">
      <button className="bb-hamburger" onClick={() => setSidebarOpen(!sidebarOpen)} aria-label="Toggle navigation">
        <svg width="20" height="20" viewBox="0 0 20 20" fill="currentColor"><path d="M2 4.5h16V6H2V4.5Zm0 4.75h16v1.5H2v-1.5ZM2 14h16v1.5H2V14Z" /></svg>
      </button>
      <div className="bb-product-mark" aria-hidden="true"><span /><span /></div>
      <div className="bb-product-name">Agent Ledger</div>
      <div className="bb-product-divider" />
      <div className="bb-product-context">Project workspace</div>
    </header>
    <div className="bb-container">
      <aside className={`bb-sidebar ${sidebarOpen ? 'open' : ''}`}>
        <div className="bb-repo-identity"><div className="bb-repo-avatar">AL</div><div><strong>Agent Ledger</strong><span>Repository</span></div></div>
        <div className="bb-sidebar-label">Workspace</div>
        <nav className="bb-sidebar-nav" aria-label="Workspace navigation">
          {navItems.map(([view, label, path]) => <button key={view} className={`bb-nav-item ${currentView === view ? 'active' : ''}`} onClick={() => navigate(view)}><svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">{path}</svg><span>{label}</span></button>)}
        </nav>
      </aside>
      {sidebarOpen && <button className="bb-sidebar-scrim" onClick={() => setSidebarOpen(false)} aria-label="Close navigation" />}
      <main className="bb-main">{children}</main>
    </div>
  </div>
}
