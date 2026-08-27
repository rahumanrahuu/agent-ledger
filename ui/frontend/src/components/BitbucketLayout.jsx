import React, { useState } from 'react'
import './BitbucketLayout.css'

export default function BitbucketLayout({ children, onNavigate, currentView }) {
  const [sidebarOpen, setSidebarOpen] = useState(true)

  const handleNavClick = (view) => {
    if (onNavigate) {
      onNavigate(view)
    }
  }

  return (
    <div className="bb-layout">
      {/* Top Navigation */}
      <nav className="bb-topnav">
        <div className="bb-topnav-left">
          <button
            className="bb-hamburger"
            onClick={() => setSidebarOpen(!sidebarOpen)}
            aria-label="Toggle sidebar"
          >
            <svg width="20" height="20" viewBox="0 0 20 20" fill="currentColor">
              <path d="M2 4h16v2H2V4zm0 5h16v2H2V9zm0 5h16v2H2v-2z" />
            </svg>
          </button>

          <div className="bb-logo">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="currentColor">
              <path d="M2 2h20v20H2V2zm4 4h4v4H6V6zm6 0h4v4h-4V6zM6 12h4v4H6v-4zm6 0h4v4h-4v-4z" />
            </svg>
            <span>Agent Ledger</span>
          </div>
        </div>

        <div className="bb-topnav-right">
          <button className="bb-icon-btn" aria-label="Search">
            <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor">
              <circle cx="8" cy="8" r="6" />
              <path d="M12 12l4 4" />
            </svg>
          </button>
          <button className="bb-icon-btn" aria-label="Notifications">
            <svg width="20" height="20" viewBox="0 0 20 20" fill="currentColor">
              <path d="M10 18c.5 0 1-.5 1-1h-2c0 .5.5 1 1 1zm6-6v-3c0-3-2-5.5-5-6v-1h-2v1c-3 .5-5 3-5 6v3l-2 2v1h14v-1l-2-2z" />
            </svg>
          </button>
          <button className="bb-icon-btn" aria-label="Settings">
            <svg width="20" height="20" viewBox="0 0 20 20" fill="currentColor">
              <circle cx="10" cy="10" r="2" />
              <path d="M10 2c.5 0 1 .5 1 1v.5c1.3.3 2.5 1 3.5 2l.3-.3c.4-.4 1-.4 1.4 0l.8.8c.4.4.4 1 0 1.4l-.3.3c1 1 1.7 2.2 2 3.5h.5c.5 0 1 .5 1 1v1.2c0 .5-.5 1-1 1h-.5c-.3 1.3-1 2.5-2 3.5l.3.3c.4.4.4 1 0 1.4l-.8.8c-.4.4-1 .4-1.4 0l-.3-.3c-1 1-2.2 1.7-3.5 2v.5c0 .5-.5 1-1 1h-1.2c-.5 0-1-.5-1-1v-.5c-1.3-.3-2.5-1-3.5-2l-.3.3c-.4.4-1 .4-1.4 0l-.8-.8c-.4-.4-.4-1 0-1.4l.3-.3c-1-1-1.7-2.2-2-3.5h-.5c-.5 0-1-.5-1-1v-1.2c0-.5.5-1 1-1h.5c.3-1.3 1-2.5 2-3.5l-.3-.3c-.4-.4-.4-1 0-1.4l.8-.8c.4-.4 1-.4 1.4 0l.3.3c1-1 2.2-1.7 3.5-2v-.5c0-.5.5-1 1-1h1.2z" />
            </svg>
          </button>
          <div className="bb-user-menu">
            <img src="https://avatar.example.com/user" alt="User" className="bb-avatar" />
          </div>
        </div>
      </nav>

      <div className="bb-container">
        {/* Sidebar */}
        <aside className={`bb-sidebar ${sidebarOpen ? 'open' : 'closed'}`}>
          <div className="bb-sidebar-content">
            <div className="bb-sidebar-section">
              <div className="bb-sidebar-section-title">Repository</div>
              <nav className="bb-sidebar-nav">
                <button
                  className={`bb-nav-item ${currentView === 'overview' ? 'active' : ''}`}
                  onClick={() => handleNavClick('overview')}
                >
                  <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                    <path d="M2 2h4v4H2V2zm0 6h4v4H2v-4zm6-6h4v4h-4V2zm0 6h4v4h-4v-4z" />
                  </svg>
                  <span>Overview</span>
                </button>
                <button
                  className={`bb-nav-item ${currentView === 'memories' ? 'active' : ''}`}
                  onClick={() => handleNavClick('memories')}
                >
                  <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                    <path d="M2 2h12v12H2V2zm1 1h10v10H3V3z" />
                  </svg>
                  <span>Memories</span>
                </button>
                <button
                  className={`bb-nav-item ${currentView === 'timeline' ? 'active' : ''}`}
                  onClick={() => handleNavClick('timeline')}
                >
                  <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                    <path d="M2 2v12h12V2H2zm1 1h10v10H3V3zm2 2h6v1H5V5zm0 2h6v1H5V7zm0 2h4v1H5V9z" />
                  </svg>
                  <span>Timeline</span>
                </button>
                <button
                  className={`bb-nav-item ${currentView === 'sessions' ? 'active' : ''}`}
                  onClick={() => handleNavClick('sessions')}
                >
                  <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                    <path d="M8 1c1.657 0 3 1.343 3 3s-1.343 3-3 3-3-1.343-3-3 1.343-3 3-3zm0 7c2.757 0 5 1.343 5 3v3H3v-3c0-1.657 2.243-3 5-3z" />
                  </svg>
                  <span>Sessions</span>
                </button>
                <button
                  className={`bb-nav-item ${currentView === 'graph' ? 'active' : ''}`}
                  onClick={() => handleNavClick('graph')}
                >
                  <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                    <path d="M5 4c1.1 0 2-.9 2-2s-.9-2-2-2-2 .9-2 2 .9 2 2 2zm0 2c-1.66 0-5 .84-5 2.5V12h10v-3.5C10 6.84 6.66 6 5 6zm11-6h-3v3h-2V0h-3v3h-2V0H6v3h3v2h2V3h3V0zm-2 8v3h-3v-3h3z" />
                  </svg>
                  <span>Graph</span>
                </button>
              </nav>
            </div>

            <div className="bb-sidebar-section">
              <div className="bb-sidebar-section-title">Tools</div>
              <nav className="bb-sidebar-nav">
                <button className="bb-nav-item">
                  <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                    <circle cx="7" cy="7" r="5" fill="none" stroke="currentColor" strokeWidth="2" />
                    <path d="M11 11l3 3" stroke="currentColor" strokeWidth="2" />
                  </svg>
                  <span>Search</span>
                </button>
                <button className="bb-nav-item">
                  <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                    <path d="M14 8c0-3.314-2.686-6-6-6S2 4.686 2 8s2.686 6 6 6 6-2.686 6-6zM8 3v8M5 8h6" stroke="currentColor" strokeWidth="1.5" fill="none" />
                  </svg>
                  <span>Constraints</span>
                </button>
                <button className="bb-nav-item">
                  <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                    <circle cx="8" cy="8" r="1.5" />
                    <path d="M8 1v2M8 13v2M1 8h2M13 8h2M2.6 2.6l1.4 1.4M11.8 11.8l1.4 1.4M2.6 13.4l1.4-1.4M11.8 4.2l1.4-1.4" stroke="currentColor" strokeWidth="1.5" fill="none" />
                  </svg>
                  <span>Settings</span>
                </button>
              </nav>
            </div>
          </div>
        </aside>

        {/* Main Content */}
        <main className="bb-main">
          {children}
        </main>
      </div>
    </div>
  )
}
