import { NavLink } from 'react-router-dom';

const navItems = [
  {
    to: '/overview',
    label: 'Overview',
    icon: (
      <svg width="15" height="15" viewBox="0 0 15 15" fill="none">
        <rect x="1" y="1" width="5.5" height="5.5" rx="1.5" stroke="currentColor" strokeWidth="1.3"/>
        <rect x="8.5" y="1" width="5.5" height="5.5" rx="1.5" stroke="currentColor" strokeWidth="1.3"/>
        <rect x="1" y="8.5" width="5.5" height="5.5" rx="1.5" stroke="currentColor" strokeWidth="1.3"/>
        <rect x="8.5" y="8.5" width="5.5" height="5.5" rx="1.5" stroke="currentColor" strokeWidth="1.3"/>
      </svg>
    ),
  },
  {
    to: '/memories',
    label: 'Memories',
    icon: (
      <svg width="15" height="15" viewBox="0 0 15 15" fill="none">
        <path d="M7.5 1.5C5.02 1.5 3 3.52 3 6c0 1.5.72 2.84 1.83 3.69L4.5 12.5h6l-.33-2.81A4.5 4.5 0 0 0 12 6c0-2.48-2.02-4.5-4.5-4.5Z" stroke="currentColor" strokeWidth="1.3" strokeLinejoin="round"/>
        <path d="M5.5 12.5h4v1a1 1 0 0 1-1 1h-2a1 1 0 0 1-1-1v-1Z" stroke="currentColor" strokeWidth="1.3"/>
      </svg>
    ),
  },
  {
    to: '/timeline',
    label: 'Timeline',
    icon: (
      <svg width="15" height="15" viewBox="0 0 15 15" fill="none">
        <line x1="3.5" y1="1.5" x2="3.5" y2="13.5" stroke="currentColor" strokeWidth="1.3" strokeLinecap="round"/>
        <circle cx="3.5" cy="4" r="2" stroke="currentColor" strokeWidth="1.3"/>
        <circle cx="3.5" cy="9" r="2" stroke="currentColor" strokeWidth="1.3"/>
        <line x1="6.5" y1="4" x2="13" y2="4" stroke="currentColor" strokeWidth="1.3" strokeLinecap="round"/>
        <line x1="6.5" y1="9" x2="11.5" y2="9" stroke="currentColor" strokeWidth="1.3" strokeLinecap="round"/>
      </svg>
    ),
  },
  {
    to: '/sessions',
    label: 'Sessions',
    icon: (
      <svg width="15" height="15" viewBox="0 0 15 15" fill="none">
        <rect x="1" y="2.5" width="13" height="10" rx="1.5" stroke="currentColor" strokeWidth="1.3"/>
        <path d="M4.5 6.5h6M4.5 9.5h4" stroke="currentColor" strokeWidth="1.3" strokeLinecap="round"/>
      </svg>
    ),
  },
  {
    to: '/knowledge-graph',
    label: 'Knowledge Graph',
    icon: (
      <svg width="15" height="15" viewBox="0 0 15 15" fill="none">
        <circle cx="7.5" cy="7.5" r="2" stroke="currentColor" strokeWidth="1.3"/>
        <circle cx="2.5" cy="3" r="1.5" stroke="currentColor" strokeWidth="1.2"/>
        <circle cx="12.5" cy="3" r="1.5" stroke="currentColor" strokeWidth="1.2"/>
        <circle cx="2.5" cy="12" r="1.5" stroke="currentColor" strokeWidth="1.2"/>
        <circle cx="12.5" cy="12" r="1.5" stroke="currentColor" strokeWidth="1.2"/>
        <line x1="4" y1="3.8" x2="6.2" y2="6.2" stroke="currentColor" strokeWidth="1.2"/>
        <line x1="11" y1="3.8" x2="8.8" y2="6.2" stroke="currentColor" strokeWidth="1.2"/>
        <line x1="4" y1="11.2" x2="6.2" y2="8.8" stroke="currentColor" strokeWidth="1.2"/>
        <line x1="11" y1="11.2" x2="8.8" y2="8.8" stroke="currentColor" strokeWidth="1.2"/>
      </svg>
    ),
  },
];

interface SidebarProps {
  projectName?: string;
  version?: string;
}

export default function Sidebar({ projectName, version }: SidebarProps) {
  return (
    <aside className="w-56 shrink-0 h-full bg-card border-r border-border flex flex-col">
      <div className="px-4 h-14 flex items-center gap-2.5 border-b border-border">
        <div className="w-7 h-7 rounded-lg bg-primary flex items-center justify-center shrink-0">
          <svg width="13" height="13" viewBox="0 0 14 14" fill="none">
            <path d="M7 1.5L2.5 4v6l4.5 2.5 4.5-2.5V4L7 1.5Z" stroke="white" strokeWidth="1.3" strokeLinejoin="round"/>
            <path d="M7 1.5v11M2.5 4l4.5 2.5 4.5-2.5" stroke="white" strokeWidth="1.1" strokeLinejoin="round"/>
          </svg>
        </div>
        <span className="font-semibold text-sm text-foreground tracking-tight">Agent Ledger</span>
      </div>

      <nav className="flex-1 py-3 px-2 overflow-y-auto">
        <div className="space-y-0.5">
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              className={({ isActive }) =>
                `w-full flex items-center gap-2.5 px-3 py-2 rounded-md text-sm font-medium transition-colors duration-100 ${
                  isActive
                    ? 'bg-primary-light text-primary'
                    : 'text-muted-foreground hover:bg-muted hover:text-foreground'
                }`
              }
            >
              <span className="shrink-0">{item.icon}</span>
              <span>{item.label}</span>
            </NavLink>
          ))}
        </div>
      </nav>

      <div className="p-3 border-t border-border">
        <div className="px-1 py-1 min-w-0">
          <p className="text-xs font-medium text-foreground truncate">{projectName || 'Agent Ledger'}</p>
          <p className="text-[11px] text-muted-foreground mt-0.5 font-mono truncate">
            {version ? `Backend ${version}` : 'Local workspace'}
          </p>
        </div>
      </div>
    </aside>
  );
}
