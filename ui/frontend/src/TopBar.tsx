import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { search as apiSearch } from './api/client';
import type { SearchResult } from './api/types';
import type { WsStatus } from './api/client';
import { formatRelative } from './components/ui';

interface TopBarProps {
  projectName?: string;
  branch?: string;
  commit?: string;
  lastActivity?: string;
  wsStatus: WsStatus;
  onOpenCommandPalette?: () => void;
}

const wsLabels: Record<WsStatus, string> = {
  connected: 'Live',
  connecting: 'Connecting…',
  reconnecting: 'Reconnecting…',
  disconnected: 'Offline',
};

const wsColors: Record<WsStatus, string> = {
  connected: 'bg-success',
  connecting: 'bg-warning',
  reconnecting: 'bg-warning',
  disconnected: 'bg-muted-foreground',
};

const wsTextColors: Record<WsStatus, string> = {
  connected: 'text-success',
  connecting: 'text-warning',
  reconnecting: 'text-warning',
  disconnected: 'text-muted-foreground',
};

export default function TopBar({ projectName, branch, commit, lastActivity, wsStatus, onOpenCommandPalette }: TopBarProps) {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<SearchResult[]>([]);
  const [searching, setSearching] = useState(false);
  const [showResults, setShowResults] = useState(false);
  const navigate = useNavigate();

  async function handleSearch(q: string) {
    setQuery(q);
    if (!q.trim()) { setResults([]); setShowResults(false); return; }
    setSearching(true);
    setShowResults(true);
    try {
      const res = await apiSearch(q);
      setResults(res);
    } finally {
      setSearching(false);
    }
  }

  function handleResultClick(r: SearchResult) {
    setShowResults(false);
    setQuery('');
    if (r.type === 'session') navigate(`/sessions/${r.id}`);
    else if (['decision', 'discovery', 'failure', 'constraint', 'checkpoint'].includes(r.type))
      navigate(`/timeline?event=${r.id}`);
    else if (['fact', 'rule', 'pattern', 'entity', 'preference', 'insight'].includes(r.type))
      navigate(`/memories?id=${r.id}`);
  }

  return (
    <header className="h-14 border-b border-border bg-card flex items-center px-5 gap-4 shrink-0 relative z-20">
      {/* Project context */}
      <div className="flex items-center gap-2 min-w-0 shrink-0">
        {projectName && (
          <span className="text-sm font-semibold text-foreground truncate max-w-36">{projectName}</span>
        )}
        {branch && (
          <div className="flex items-center gap-1.5 text-xs text-muted-foreground font-mono bg-muted px-2 py-1 rounded-md">
            <svg width="9" height="10" viewBox="0 0 10 12" fill="none" className="shrink-0">
              <circle cx="3" cy="2" r="1.5" stroke="currentColor" strokeWidth="1.2"/>
              <circle cx="3" cy="10" r="1.5" stroke="currentColor" strokeWidth="1.2"/>
              <circle cx="8" cy="5" r="1.5" stroke="currentColor" strokeWidth="1.2"/>
              <path d="M3 3.5V8.5M3 3.5C3 3.5 3 5 8 5" stroke="currentColor" strokeWidth="1.2" strokeLinecap="round"/>
            </svg>
            {branch}
          </div>
        )}
        {commit && (
          <span className="text-xs font-mono text-muted-foreground bg-muted px-2 py-1 rounded-md hidden md:block">
            {commit.slice(0, 8)}
          </span>
        )}
      </div>

      {/* Search */}
      <div className="flex-1 max-w-md mx-auto relative">
        <div className="flex items-center gap-2 h-8 px-3 rounded-lg border border-border bg-muted focus-within:border-primary focus-within:bg-card transition-colors">
          <svg width="12" height="12" viewBox="0 0 14 14" fill="none" className="shrink-0 text-muted-foreground">
            <circle cx="6" cy="6" r="4.5" stroke="currentColor" strokeWidth="1.4"/>
            <path d="m9.5 9.5 2.5 2.5" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round"/>
          </svg>
          <input
            type="text"
            placeholder="Search sessions, events, memories…"
            value={query}
            onChange={(e) => handleSearch(e.target.value)}
            onFocus={() => query && setShowResults(true)}
            onBlur={() => setTimeout(() => setShowResults(false), 150)}
            className="flex-1 bg-transparent text-xs text-foreground placeholder:text-muted-foreground outline-none"
          />
          {!query && (
            <button
              onClick={onOpenCommandPalette}
              title="Open Command Palette (⌘K)"
              className="text-xs text-muted-foreground hover:text-foreground font-mono bg-muted-foreground/10 px-1.5 py-0.5 rounded border border-border hidden lg:block cursor-pointer transition-colors"
            >
              ⌘K
            </button>
          )}
          {query && (
            <button onClick={() => { setQuery(''); setResults([]); setShowResults(false); }}
              className="text-muted-foreground hover:text-foreground">
              <svg width="10" height="10" viewBox="0 0 10 10" fill="none">
                <path d="M1 1l8 8M9 1L1 9" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round"/>
              </svg>
            </button>
          )}
        </div>

        {/* Search dropdown */}
        {showResults && (
          <div className="absolute top-full left-0 right-0 mt-1 bg-card border border-border rounded-xl shadow-lg overflow-hidden">
            {searching && (
              <div className="px-4 py-3 text-xs text-muted-foreground">Searching…</div>
            )}
            {!searching && results.length === 0 && query && (
              <div className="px-4 py-3 text-xs text-muted-foreground">No results for "{query}"</div>
            )}
            {!searching && results.map((r) => (
              <button
                key={r.id}
                onMouseDown={() => handleResultClick(r)}
                className="w-full text-left px-4 py-2.5 hover:bg-muted transition-colors border-b border-border last:border-0"
              >
                <div className="flex items-center gap-2 mb-0.5">
                  <span className="text-xs text-muted-foreground capitalize">{r.type}</span>
                  {r.sessionName && (
                    <span className="text-xs text-muted-foreground">· {r.sessionName}</span>
                  )}
                  <span className="ml-auto text-xs text-muted-foreground">{formatRelative(r.timestamp)}</span>
                </div>
                <p className="text-xs font-medium text-foreground truncate">{r.title}</p>
              </button>
            ))}
          </div>
        )}
      </div>

      {/* Right */}
      <div className="flex items-center gap-3 ml-auto shrink-0">
        {lastActivity && (
          <span className="text-xs text-muted-foreground hidden xl:block">Updated {formatRelative(lastActivity)}</span>
        )}
        <div className="flex items-center gap-1.5 px-2.5 py-1.5 rounded-md border border-border bg-card">
          <div className={`w-1.5 h-1.5 rounded-full ${wsColors[wsStatus]} ${wsStatus === 'connected' ? 'animate-pulse' : ''}`} />
          <span className={`text-xs font-medium ${wsTextColors[wsStatus]}`}>{wsLabels[wsStatus]}</span>
        </div>
      </div>
    </header>
  );
}
